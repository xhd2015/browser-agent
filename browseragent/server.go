package browseragent

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

type controlServer struct {
	sess *session
}

func newControlServer(sess *session) *controlServer {
	return &controlServer{sess: sess}
}

func (c *controlServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", c.handleHealth)
	mux.HandleFunc("/v1/session", c.handleSession)
	mux.HandleFunc("/v1/jobs", c.handleJobs)
	mux.HandleFunc("/v1/ws", c.handleWS)
	mux.HandleFunc("/go", c.handleGo)
	mux.HandleFunc("/assets/", c.handleAssets)
	mux.HandleFunc("/", c.handleRoot)
	return mux
}

func (c *controlServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (c *controlServer) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("session")
	// Single-session serve: empty session query uses live session.
	if id != "" && id != c.sess.id {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "session not found",
			"message": "unknown session id",
		})
		return
	}
	snap := c.sess.snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

type jobsRequest struct {
	SessionID string         `json:"session_id"`
	Type      string         `json:"type"`
	Params    map[string]any `json:"params"`
	TimeoutMS int64          `json:"timeout_ms"`
}

func (c *controlServer) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req jobsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json: " + err.Error()})
		return
	}
	sid := req.SessionID
	if sid == "" {
		sid = r.URL.Query().Get("session")
	}
	if sid != "" && sid != c.sess.id {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "session not found",
			"message": "unknown session id",
		})
		return
	}

	timeoutMS := req.TimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 30000
	}
	jobType := req.Type
	if jobType == "" {
		jobType = "eval"
	}

	enqueued, err := c.sess.queue.Enqueue(Job{
		SessionID: c.sess.id,
		Type:      jobType,
		Params:    req.Params,
		TimeoutMS: timeoutMS,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Push to extension if connected.
	_ = c.pushJob(enqueued)

	// Hold until result / timeout / disconnect-fail.
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()
	res, waitErr := c.sess.queue.Wait(ctx, enqueued.ID)
	if waitErr != nil && res.Error == "" {
		// Ensure timeout wording.
		if ctx.Err() != nil {
			res = JobResult{
				JobID: enqueued.ID,
				OK:    false,
				Error: "timeout waiting for job result",
			}
		} else {
			res = JobResult{
				JobID: enqueued.ID,
				OK:    false,
				Error: waitErr.Error(),
			}
		}
	}
	if res.JobID == "" {
		res.JobID = enqueued.ID
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

func (c *controlServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// Let /assets/ and other registered routes be handled by their own
		// patterns; anything else is 404.
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			c.handleAssets(w, r)
			return
		}
		http.NotFound(w, r)
		return
	}
	c.handleGo(w, r)
}

func (c *controlServer) handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// URL path: /assets/session-page.js → embed: embedded/session-page/assets/session-page.js
	rel := strings.TrimPrefix(r.URL.Path, "/assets/")
	rel = path.Clean("/" + rel)
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || rel == "." || strings.Contains(rel, "..") {
		http.NotFound(w, r)
		return
	}
	embedPath := embeddedSessionPageRoot + "/assets/" + rel
	data, err := fs.ReadFile(embeddedSessionPage, embedPath)
	if err != nil {
		// Also try without "assets/" double-nest if path already includes it.
		alt := embeddedSessionPageRoot + "/" + rel
		data, err = fs.ReadFile(embeddedSessionPage, alt)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	ct := contentTypeForAsset(rel)
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

func contentTypeForAsset(name string) string {
	low := strings.ToLower(name)
	switch {
	case strings.HasSuffix(low, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(low, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(low, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(low, ".json"):
		return "application/json; charset=utf-8"
	case strings.HasSuffix(low, ".map"):
		return "application/json; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

func (c *controlServer) handleGo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = c.sess.id
	}
	snap := c.sess.snapshot()
	// Prefer embedded session-page fixture when present.
	if htmlBody, err := readEmbeddedSessionIndex(); err == nil && strings.TrimSpace(htmlBody) != "" {
		out := injectSessionBoot(htmlBody, sessionID, snap)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(out))
		return
	}
	// Fallback: pure Go HTML shell (keeps prior SPA markers).
	c.writeFallbackSessionHTML(w, sessionID, snap)
}

// isSessionSPAHTML reports whether html is a Vite/React session shell that will
// mount its own status / identity / install UI. When true, injectSessionBoot
// must not inject full SSR panels (would duplicate React SessionPageApp).
func isSessionSPAHTML(htmlBody string) bool {
	low := strings.ToLower(htmlBody)
	// Production vite multi-page entry.
	if strings.Contains(low, "assets/session-page") {
		return true
	}
	// Module script + root mount (SPA fixture / production).
	if strings.Contains(low, `type="module"`) &&
		(strings.Contains(low, `id="root"`) || strings.Contains(low, "data-browser-agent-root")) {
		return true
	}
	return false
}

// injectSessionBoot injects boot JSON / data attrs / __BROWSER_AGENT for the
// live session. For SPA shells: boot only (React owns status/identity/install).
// For non-SPA shells: also inject SSR identity + install panels once.
func injectSessionBoot(htmlBody, sessionID string, snap sessionSnapshot) string {
	esc := html.EscapeString(sessionID)
	bootJSON := FormatSessionBootJSON(sessionID)
	// Escape </script> in JSON if session id ever contained that sequence.
	bootJSONSafe := strings.ReplaceAll(bootJSON, "</", "<\\/")

	bootBlock := fmt.Sprintf(`<script type="application/json" id="browser-agent-boot">%s</script>
<script>
  window.__BROWSER_AGENT = {
    product: "browser-agent",
    controlPort: 43761,
    defaultAddr: "127.0.0.1:43761",
    sessionId: %q
  };
</script>
`, bootJSONSafe, sessionID)

	// Inject into <head> when present; otherwise prepend after <html>.
	out := htmlBody
	if idx := strings.Index(strings.ToLower(out), "</head>"); idx >= 0 {
		out = out[:idx] + bootBlock + out[idx:]
	} else if idx := strings.Index(strings.ToLower(out), "<head>"); idx >= 0 {
		// after <head>
		end := idx + len("<head>")
		// handle <head ...>
		if gt := strings.Index(out[idx:], ">"); gt >= 0 {
			end = idx + gt + 1
		}
		out = out[:end] + "\n" + bootBlock + out[end:]
	} else {
		out = bootBlock + out
	}

	// Ensure body carries data-session-id / data-control-port / data-product.
	bodyAttrs := fmt.Sprintf(
		` data-browser-agent-session data-session-id="%s" data-product="browser-agent" data-control-port="43761"`,
		esc,
	)
	low := strings.ToLower(out)
	if bi := strings.Index(low, "<body"); bi >= 0 {
		gt := strings.Index(out[bi:], ">")
		if gt >= 0 {
			// Insert attrs before closing > of <body ...>
			insertAt := bi + gt
			// Avoid duplicating if already present.
			tag := out[bi:insertAt]
			if !strings.Contains(strings.ToLower(tag), "data-session-id") {
				out = out[:insertAt] + bodyAttrs + out[insertAt:]
			}
		}
	}

	// Ensure /v1/session is referenced for SPA poll contracts (fixture JS has it;
	// add a noscript/data marker if somehow missing after injection).
	if !strings.Contains(out, "/v1/session") {
		marker := fmt.Sprintf(
			`<script>/* session poll */ fetch('/v1/session?session='+encodeURIComponent(%q));</script>`,
			sessionID,
		)
		if idx := strings.Index(strings.ToLower(out), "</body>"); idx >= 0 {
			out = out[:idx] + marker + out[idx:]
		} else {
			out += marker
		}
	}

	// SPA path: React SessionPageApp owns visible status/identity/install UI.
	// Inject only a non-visible install marker so static HTML still documents
	// chrome://extensions for contracts/tests (no second visible panel).
	if isSessionSPAHTML(htmlBody) {
		if !strings.Contains(out, "data-browser-agent-install") {
			marker := `<meta name="browser-agent-install" content="chrome://extensions Load unpacked" data-browser-agent-install data-install-via-spa="1" />`
			if idx := strings.Index(strings.ToLower(out), "</head>"); idx >= 0 {
				out = out[:idx] + marker + out[idx:]
			} else {
				out += marker
			}
		}
		return out
	}

	// Non-SPA fallback shell: inject identity + install once if missing.
	if !strings.Contains(out, "data-browser-agent-ext-identity") &&
		!strings.Contains(out, "browser-agent-ext-identity") {
		panel := extensionIdentityPanelHTML(sessionID, snap)
		if idx := strings.Index(strings.ToLower(out), "</body>"); idx >= 0 {
			out = out[:idx] + panel + out[idx:]
		} else {
			out += panel
		}
	}

	// Identity panel may mention "Load unpacked"; only skip install inject when
	// the dedicated install panel marker is already present.
	if !strings.Contains(out, "data-browser-agent-install") &&
		!strings.Contains(out, `id="browser-agent-install"`) {
		panel := `
<details id="browser-agent-install" open data-browser-agent-install data-product="browser-agent" data-control-port="43761">
  <summary>Install browser-agent extension</summary>
  <div>
    <p>Load the unpacked Chrome extension that connects to <code>127.0.0.1:43761</code>.</p>
    <ol>
      <li>Open chrome://extensions</li>
      <li>Enable Developer mode</li>
      <li>Load unpacked the browser-agent package</li>
      <li>Keep this page open so the extension can attach to the session</li>
    </ol>
    <p class="muted">Or run: <code>browser-agent install-chrome-extension</code></p>
  </div>
</details>
`
		if idx := strings.Index(strings.ToLower(out), "</body>"); idx >= 0 {
			out = out[:idx] + panel + out[idx:]
		} else {
			out += panel
		}
	}

	return out
}

// extensionIdentityPanelHTML is SSR UI for bundled vs loaded extension package
// identity (version + md5). Polls /v1/session so SPA shells get it without rebuild.
func extensionIdentityPanelHTML(sessionID string, snap sessionSnapshot) string {
	bundledVer := html.EscapeString(snap.BundledExtension.Version)
	bundledMD5 := html.EscapeString(snap.BundledExtension.MD5)
	loadedVer := html.EscapeString(snap.Extension.Version)
	loadedMD5 := html.EscapeString(snap.Extension.BundleMD5)
	match := html.EscapeString(snap.ExtensionMatch)
	if bundledVer == "" {
		bundledVer = "—"
	}
	if bundledMD5 == "" {
		bundledMD5 = "—"
	}
	if loadedVer == "" {
		loadedVer = "—"
	}
	if loadedMD5 == "" {
		loadedMD5 = "—"
	}
	if match == "" {
		match = ExtensionMatchNotConnected
	}
	path := html.EscapeString(snap.ExtensionInstallPath)
	return fmt.Sprintf(`
<style>
  #browser-agent-ext-identity {
    border: 1px solid #ccc; border-radius: 8px; padding: 0.75rem 1rem; margin: 1rem 0;
    font-family: system-ui, sans-serif; font-size: 0.95rem;
  }
  #browser-agent-ext-identity h2 { font-size: 1.05rem; margin: 0 0 0.5rem; }
  #browser-agent-ext-identity code { background: #f4f4f4; padding: 0.1em 0.35em; word-break: break-all; }
  #browser-agent-ext-identity .ba-id-row { margin: 0.25rem 0; }
  #browser-agent-ext-identity .ba-match-ok { color: #0a7a2f; font-weight: 600; }
  #browser-agent-ext-identity .ba-match-warn { color: #c45c00; font-weight: 600; }
  #browser-agent-ext-identity .muted { color: #666; font-size: 0.85rem; }
</style>
<section id="browser-agent-ext-identity" data-browser-agent-ext-identity>
  <h2>Extension package</h2>
  <div class="ba-id-row"><strong>Bundled (this serve)</strong>
    version <code id="ba-bundled-ver">%s</code>
    md5 <code id="ba-bundled-md5">%s</code>
  </div>
  <div class="ba-id-row"><strong>Loaded (Chrome)</strong>
    version <code id="ba-loaded-ver">%s</code>
    md5 <code id="ba-loaded-md5">%s</code>
  </div>
  <div class="ba-id-row"><strong>Match:</strong> <span id="ba-ext-match" class="%s">%s</span></div>
  <p class="muted" id="ba-ext-path">Load unpacked: <code>%s</code></p>
</section>
<script>
(function() {
  var sessionId = %q;
  function $(id) { return document.getElementById(id); }
  function dash(v) { return (v && String(v).length) ? String(v) : "—"; }
  function matchClass(m) {
    if (m === "ok") return "ba-match-ok";
    if (m === "not_connected") return "muted";
    return "ba-match-warn";
  }
  function render(data) {
    if (!data) return;
    var b = data.bundled_extension || {};
    var e = data.extension || {};
    var m = data.extension_match || "not_connected";
    if ($("ba-bundled-ver")) $("ba-bundled-ver").textContent = dash(b.version);
    if ($("ba-bundled-md5")) $("ba-bundled-md5").textContent = dash(b.md5);
    if ($("ba-loaded-ver")) $("ba-loaded-ver").textContent = e.connected ? dash(e.version) : "—";
    if ($("ba-loaded-md5")) $("ba-loaded-md5").textContent = e.connected ? dash(e.bundle_md5) : "—";
    var el = $("ba-ext-match");
    if (el) {
      el.textContent = m;
      el.className = matchClass(m);
    }
    if ($("ba-ext-path") && (data.extension_install_path || (b && b.path))) {
      var p = data.extension_install_path || b.path || "";
      $("ba-ext-path").innerHTML = "Load unpacked: <code>" + p.replace(/</g, "&lt;") + "</code>";
    }
  }
  function poll() {
    fetch("/v1/session?session=" + encodeURIComponent(sessionId))
      .then(function(r) { return r.json(); })
      .then(render)
      .catch(function() {});
  }
  poll();
  setInterval(poll, 1000);
})();
</script>
`, bundledVer, bundledMD5, loadedVer, loadedMD5, matchClass(match), match, path, sessionID)
}

func matchClass(match string) string {
	switch match {
	case ExtensionMatchOK:
		return "ba-match-ok"
	case ExtensionMatchNotConnected:
		return "muted"
	default:
		return "ba-match-warn"
	}
}

func (c *controlServer) writeFallbackSessionHTML(w http.ResponseWriter, sessionID string, snap sessionSnapshot) {
	esc := html.EscapeString(sessionID)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	identity := extensionIdentityPanelHTML(sessionID, snap)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>browser-agent session</title>
<style>
  body { font-family: system-ui, sans-serif; margin: 1.5rem; max-width: 42rem; }
  code, pre.path { background: #f4f4f4; padding: 0.1em 0.3em; }
  pre.path { padding: 0.6rem 0.8rem; overflow-x: auto; white-space: pre-wrap; word-break: break-all; }
  #browser-agent-status { border: 1px solid #ccc; border-radius: 8px; padding: 1rem; margin-top: 1rem; }
  #browser-agent-install { border: 1px solid #c9a227; background: #fffbeb; border-radius: 8px; padding: 0.75rem 1rem; margin-top: 1rem; }
  .hint { color: #444; margin-top: 0.75rem; }
  .muted { color: #888; font-size: 0.9rem; }
</style>
<script type="application/json" id="browser-agent-boot">%s</script>
<script>
  window.__BROWSER_AGENT = {
    product: "browser-agent",
    controlPort: 43761,
    defaultAddr: "127.0.0.1:43761",
    sessionId: %q
  };
</script>
</head>
<body data-browser-agent-session data-session-id="%s" data-product="browser-agent" data-control-port="43761">
  <div id="root" data-browser-agent-root class="browser-agent-root">
  <h1>browser-agent</h1>
  <p>Session <code id="session-id">%s</code> is active.</p>
  <p class="muted">Control port <strong>43761</strong> · product <code>browser-agent</code></p>
  <div id="browser-agent-status" data-browser-agent-status>
    <div><strong>Phase:</strong> <span id="st-phase">…</span></div>
    <div><strong>Extension:</strong> <span id="st-ext">…</span></div>
    <div class="hint" id="st-hint">Loading status…</div>
  </div>
  %s
  <details id="browser-agent-install" open data-browser-agent-install>
    <summary>Install browser-agent extension</summary>
    <div>
      <p>Load the unpacked Chrome extension that connects to <code>127.0.0.1:43761</code>.</p>
      <ol>
        <li>Open chrome://extensions</li>
        <li>Enable Developer mode</li>
        <li>Load unpacked the browser-agent package</li>
        <li>Keep this page open so the extension can attach to the session</li>
      </ol>
    </div>
  </details>
  <script>
(function() {
  var sessionId = %q;
  function $(id) { return document.getElementById(id); }
  function render(data) {
    if (!data) return;
    $('st-phase').textContent = data.phase || '';
    var ext = data.extension || {};
    var bits = [];
    bits.push(ext.connected ? 'connected' : 'not connected');
    if (ext.version) bits.push('v' + ext.version);
    if (ext.bundle_md5) bits.push('md5 ' + ext.bundle_md5);
    bits.push(ext.supports_browser_agent ? 'supports browser-agent' : 'no browser-agent support');
    $('st-ext').textContent = bits.join(' · ');
    $('st-hint').textContent = data.hint || '';
  }
  function poll() {
    fetch('/v1/session?session=' + encodeURIComponent(sessionId))
      .then(function(r) { return r.json().then(function(j) { return { ok: r.ok, body: j }; }); })
      .then(function(res) {
        if (res.ok) render(res.body);
        else $('st-hint').textContent = (res.body && (res.body.error || res.body.message)) || 'session error';
      })
      .catch(function() { $('st-hint').textContent = 'Waiting for control server…'; });
  }
  poll();
  setInterval(poll, 500);
})();
  </script>
  </div>
</body></html>`, FormatSessionBootJSON(sessionID), sessionID, esc, esc, identity, sessionID)
}
