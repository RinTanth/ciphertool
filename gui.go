package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
)

const guiAddr = "127.0.0.1:8765"

const guiPage = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>ciphertool 🍮</title>
<style>
  :root {
    --cream:    #FBF3E7;
    --panel:    #FFFBF4;
    --cocoa:    #4A3728;
    --cocoa-70: #6F5A47;
    --mocha:    #8C7B6B;
    --caramel:  #B98858;
    --caramel-dark: #9C6E42;
    --dust:     #EFE0C9;
    --border:   #E6D7BE;
    --error:    #B5502A;
  }
  * { box-sizing: border-box; }
  body {
    font-family: -apple-system, "SF Pro Text", "Helvetica Neue", sans-serif;
    max-width: 620px;
    margin: 48px auto;
    padding: 0 20px 60px;
    background: var(--cream);
    color: var(--cocoa);
  }
  h1 {
    font-size: 22px;
    font-weight: 600;
    letter-spacing: -0.02em;
    margin: 0 0 4px;
  }
  .subtitle {
    color: var(--mocha);
    font-size: 13px;
    margin: 0 0 28px;
  }
  .panel {
    background: var(--panel);
    border: 1px solid var(--border);
    border-radius: 18px;
    padding: 20px 20px 16px;
    margin-bottom: 20px;
    box-shadow: 0 2px 10px rgba(74, 55, 40, 0.06);
  }
  label {
    display: block;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--mocha);
    margin-bottom: 8px;
  }
  textarea {
    width: 100%;
    font-family: ui-monospace, "SF Mono", monospace;
    font-size: 13px;
    line-height: 1.5;
    padding: 12px;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: #fff;
    color: var(--cocoa);
    resize: vertical;
    outline: none;
    transition: border-color 0.15s ease;
  }
  textarea:focus { border-color: var(--caramel); }
  textarea.input { height: 76px; }
  textarea.output { height: 76px; background: var(--dust); margin-top: 14px; }
  .row {
    display: flex;
    gap: 10px;
    align-items: center;
    margin-top: 12px;
    flex-wrap: wrap;
  }
  button {
    font-family: inherit;
    font-size: 13px;
    font-weight: 600;
    padding: 8px 18px;
    border: none;
    border-radius: 999px;
    cursor: pointer;
    transition: transform 0.08s ease, background 0.15s ease;
  }
  button:active { transform: scale(0.96); }
  .btn-primary {
    background: var(--caramel);
    color: #fff;
  }
  .btn-primary:hover { background: var(--caramel-dark); }
  .btn-ghost {
    background: transparent;
    color: var(--cocoa-70);
    border: 1px solid var(--border);
  }
  .btn-ghost:hover { background: var(--dust); }
  .divider {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--mocha);
    font-size: 16px;
    margin: 4px 0;
    opacity: 0.6;
  }
  .error {
    color: var(--error);
    font-size: 12.5px;
    margin-top: 8px;
    min-height: 16px;
  }
  .copied {
    color: var(--caramel-dark);
    font-size: 12px;
    margin-left: 4px;
    opacity: 0;
    transition: opacity 0.2s ease;
  }
  .copied.show { opacity: 1; }
  input[type="text"], input[type="password"] {
    width: 100%;
    font-family: ui-monospace, "SF Mono", monospace;
    font-size: 13px;
    padding: 10px 12px;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: #fff;
    color: var(--cocoa);
    outline: none;
    transition: border-color 0.15s ease;
  }
  input[type="text"]:focus, input[type="password"]:focus { border-color: var(--caramel); }
  .key-hint {
    font-size: 12px;
    color: var(--mocha);
    margin-top: 8px;
  }
</style>
</head>
<body data-default-key="{{.DefaultKey}}">
  <h1>🍮 ciphertool</h1>
  <p class="subtitle">AES-GCM encrypt &amp; decrypt</p>

  <div class="panel">
    <label>AES-GCM Key</label>
    <input type="password" id="key" placeholder="16 / 24 / 32-byte key" oninput="saveKey()">
    <div class="row">
      <button class="btn-ghost" onclick="toggleKeyVisibility()" id="toggle-key-btn">Show</button>
      <button class="btn-ghost" onclick="resetKey()">Reset to default</button>
    </div>
    <p class="key-hint">Stored only in this browser (localStorage) — never written to disk.</p>
  </div>

  <div class="panel">
    <label>Plaintext</label>
    <textarea class="input" id="plain" placeholder="Paste text to encrypt..."></textarea>
    <div class="row">
      <button class="btn-primary" onclick="run('encrypt')">Encrypt ↓</button>
      <button class="btn-ghost" onclick="copyOut('cipher', 'copied-cipher')">Copy ciphertext</button>
      <button class="btn-ghost" onclick="formatJSON()">Format JSON</button>
      <span class="copied" id="copied-cipher">copied!</span>
    </div>
    <div class="error" id="err-encrypt"></div>
  </div>

  <div class="divider">⋮</div>

  <div class="panel">
    <label>Ciphertext (hex)</label>
    <textarea class="input" id="cipher" placeholder="Paste hex ciphertext to decrypt..."></textarea>
    <div class="row">
      <button class="btn-primary" onclick="run('decrypt')">Decrypt ↑</button>
      <button class="btn-ghost" onclick="copyOut('plain', 'copied-plain')">Copy plaintext</button>
      <span class="copied" id="copied-plain">copied!</span>
    </div>
    <div class="error" id="err-decrypt"></div>
  </div>

<script>
const DEFAULT_KEY = document.body.dataset.defaultKey;
const KEY_STORAGE = 'ciphertool_key';

(function initKey() {
  const el = document.getElementById('key');
  el.value = localStorage.getItem(KEY_STORAGE) || DEFAULT_KEY;
})();

function saveKey() {
  localStorage.setItem(KEY_STORAGE, document.getElementById('key').value);
}

function resetKey() {
  document.getElementById('key').value = DEFAULT_KEY;
  saveKey();
}

function toggleKeyVisibility() {
  const el = document.getElementById('key');
  const btn = document.getElementById('toggle-key-btn');
  const shown = el.type === 'text';
  el.type = shown ? 'password' : 'text';
  btn.textContent = shown ? 'Show' : 'Hide';
}

async function run(mode) {
  const errId = 'err-' + mode;
  document.getElementById(errId).textContent = '';
  const srcId = mode === 'encrypt' ? 'plain' : 'cipher';
  const dstId = mode === 'encrypt' ? 'cipher' : 'plain';
  const text = document.getElementById(srcId).value;
  const key = document.getElementById('key').value;
  try {
    const resp = await fetch('/api/' + mode, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({text, key}),
    });
    const data = await resp.json();
    if (!resp.ok) {
      document.getElementById(errId).textContent = data.error || 'request failed';
      return;
    }
    document.getElementById(dstId).value = data.result;
  } catch (e) {
    document.getElementById(errId).textContent = String(e);
  }
}

function formatJSON() {
  const el = document.getElementById('plain');
  const errEl = document.getElementById('err-encrypt');
  errEl.textContent = '';
  if (!el.value.trim()) return;
  try {
    el.value = JSON.stringify(JSON.parse(el.value), null, 2);
  } catch (e) {
    errEl.textContent = 'not valid JSON: ' + e.message;
  }
}

function copyOut(id, badgeId) {
  const el = document.getElementById(id);
  el.select();
  navigator.clipboard.writeText(el.value);
  const badge = document.getElementById(badgeId);
  badge.classList.add('show');
  setTimeout(() => badge.classList.remove('show'), 1200);
}
</script>
</body>
</html>`

type apiRequest struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

type apiResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// keyOrDefault lets the GUI override the built-in key per-request; falls back
// to the compiled-in aesKey when the field is left blank.
func keyOrDefault(k string) string {
	if k == "" {
		return aesKey
	}
	return k
}

func runGUI() {
	tmpl := template.Must(template.New("gui").Parse(guiPage))
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, struct{ DefaultKey string }{DefaultKey: aesKey})
	})

	mux.HandleFunc("/api/encrypt", func(w http.ResponseWriter, r *http.Request) {
		var req apiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{Error: "invalid request"})
			return
		}
		result, err := EncryptGCM(req.Text, keyOrDefault(req.Key))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiResponse{Result: result})
	})

	mux.HandleFunc("/api/decrypt", func(w http.ResponseWriter, r *http.Request) {
		var req apiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{Error: "invalid request"})
			return
		}
		result, err := DecryptGCM(req.Text, keyOrDefault(req.Key))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiResponse{Result: result})
	})

	url := "http://" + guiAddr
	fmt.Println("ciphertool GUI running at", url)
	fmt.Println("(bound to localhost only — press Ctrl+C to stop)")
	go openBrowser(url)

	if err := http.ListenAndServe(guiAddr, mux); err != nil {
		fmt.Println("server error:", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, resp apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}
