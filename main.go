package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
)

// aesKey must match frep-oc-aml's configured aes.aes_key and be 16, 24, or 32
// bytes long (AES-128/192/256). Replace the placeholder below with the real key.
const aesKey = "YOUR_AESGCM_KEY!"

// EncryptGCM and DecryptGCM mirror frep-oc-aml's api/common/aesgcm/aesgcm.go:
// AES-GCM with a random 12-byte nonce prepended to the ciphertext, hex-encoded.

func EncryptGCM(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(plaintext), nil)
	result := append(nonce, ciphertext...)

	return hex.EncodeToString(result), nil
}

func DecryptGCM(encryptedHex, key string) (string, error) {
	if encryptedHex == "" {
		return "", nil
	}

	encrypted, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(encrypted) < 12 {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce := encrypted[:12]
	ciphertext := encrypted[12:]

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	decrypted, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func main() {
	encryptText := flag.String("E", "", "plaintext to encrypt")
	decryptText := flag.String("D", "", "hex ciphertext to decrypt")
	gui := flag.Bool("gui", false, "launch a local web GUI for encrypt/decrypt")
	flag.Parse()

	switch {
	case *gui:
		// GUI lets the key be entered/overridden in the browser, so the
		// compiled-in placeholder isn't a blocker here.
		runGUI()
	case *encryptText != "":
		if aesKey == "YOUR_AESGCM_KEY!" {
			fmt.Fprintln(os.Stderr, "aesKey is still the placeholder value — edit the const in main.go, or use -gui and enter a key there")
			os.Exit(1)
		}
		result, err := EncryptGCM(*encryptText, aesKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "encrypt failed:", err)
			os.Exit(1)
		}
		fmt.Println(result)
	case *decryptText != "":
		if aesKey == "YOUR_AESGCM_KEY!" {
			fmt.Fprintln(os.Stderr, "aesKey is still the placeholder value — edit the const in main.go, or use -gui and enter a key there")
			os.Exit(1)
		}
		result, err := DecryptGCM(*decryptText, aesKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "decrypt failed:", err)
			os.Exit(1)
		}
		fmt.Println(result)
	default:
		fmt.Fprintln(os.Stderr, "usage: go run main.go -E <plaintext>   (encrypt)")
		fmt.Fprintln(os.Stderr, "       go run main.go -D <ciphertext>  (decrypt)")
		os.Exit(1)
	}
}
