# Scenario

**Feature**: Instant toolbar popup + progressive daemon / session / debugger status

```
# static: read popup.html, popup.js, background.js progressive-status contracts
Test Client -> assert shell version paints without network wait
Test Client -> assert health fetch has timeout/abort
Test Client -> assert progressive phase DOM hooks (daemon, session/ws, debugger)
Test Client -> assert SW status API or storage last-known write
Test Client -> assert popup reads last-known status from storage
```

## Preconditions

- **ModuleRoot** = workspace root (`filepath.Join(DOCTEST_ROOT, "..", "..")`).
- Classic TDD **P4**: progressive popup status. Current health-only popup is
  **obsolete** for session/WS/debugger visibility and hung-check prevention.
- Ext-source leaves: no browser; read `Chrome-Ext-Browser-Agent/public/popup.html`,
  `popup.js`, `background.js` (also accept `build/` / `src/` fallbacks).
- Scope: **Chrome Browser Agent extension only** (not browser-trace, not Firefox,
  not Capture-API popup stats tree).
- Parallel-safe: pure FS reads; no `t.Setenv` / `os.Chdir` in harness.
- MECE with:
  - `browser-agent-session-attach-gate` (P1 set/gate/leave)
  - `browser-agent-session-eager-arm` (P2 eager triggers)
  - `browser-agent-session-self-heal` (P3 boot rediscover)
  — do not re-assert those contracts here.
- Prefer new tree `browser-agent-popup-progressive-status` over extending
  `chrome-ext-popup-stats` (that tree is pure Go stats builder for Capture-API).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Grouping Setup sets `Mode = ModeExtSource`.
3. Leaf Setup sets `ExtSourceTarget`.

## Context

- Spec version **0.0.1**.
- Complements background attach/heal with **operator-visible progressive status**.
- Accept flexible implementer naming for phase ids and message types; require
  three distinct phase surfaces + timeout + last-known path.
- Optional badge is out of scope.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// extractJSFunctionBody returns the body text of `function name(...) { ... }`
// or `async function name(...) { ... }` by brace matching. ok=false if not found.
func extractJSFunctionBody(src, name string) (string, bool) {
	patterns := []string{
		"function " + name + "(",
		"async function " + name + "(",
	}
	start := -1
	for _, p := range patterns {
		idx := strings.Index(src, p)
		if idx >= 0 && (start < 0 || idx < start) {
			start = idx
		}
	}
	if start < 0 {
		for _, p := range []string{
			name + " = function",
			name + "=function",
			name + " = async function",
			name + "=async function",
		} {
			idx := strings.Index(src, p)
			if idx >= 0 && (start < 0 || idx < start) {
				start = idx
			}
		}
	}
	if start < 0 {
		return "", false
	}
	brace := strings.Index(src[start:], "{")
	if brace < 0 {
		return "", false
	}
	i := start + brace
	depth := 0
	for j := i; j < len(src); j++ {
		switch src[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return src[i+1 : j], true
			}
		}
	}
	return "", false
}

// listenerNeighborhood returns ~window bytes around the first occurrence of marker.
func listenerNeighborhood(src, marker string, window int) string {
	idx := strings.Index(src, marker)
	if idx < 0 {
		return ""
	}
	start := idx - window/4
	if start < 0 {
		start = 0
	}
	end := idx + window
	if end > len(src) {
		end = len(src)
	}
	return src[start:end]
}

// --- shell version (instant) ---

// hasInstantShellVersion — package identity paints without waiting on health fetch.
// Requires version DOM hook + sync/bundle identity source (not only post-fetch).
func hasInstantShellVersion(html, js string) bool {
	if strings.TrimSpace(html) == "" && strings.TrimSpace(js) == "" {
		return false
	}
	// DOM identity hook in HTML (preferred) or created in JS.
	hasVersionEl := strings.Contains(html, "pkg-version") ||
		strings.Contains(html, "id=\"version\"") ||
		strings.Contains(html, "id='version'") ||
		strings.Contains(html, "data-package-version") ||
		strings.Contains(html, "data-pkg-version") ||
		strings.Contains(js, "pkg-version") ||
		strings.Contains(js, "getElementById(\"version\")") ||
		strings.Contains(js, "getElementById('version')")
	if !hasVersionEl {
		return false
	}

	// Sync identity: bundle-sum script, BROWSER_AGENT_BUNDLE_VERSION, or manifest version
	// painted without depending solely on fetch().then.
	hasSyncIdentity := strings.Contains(html, "bundle-sum.js") ||
		strings.Contains(js, "BROWSER_AGENT_BUNDLE_VERSION") ||
		strings.Contains(js, "BROWSER_AGENT_BUNDLE_MD5") ||
		strings.Contains(html, "id=\"pkg-version\"") ||
		strings.Contains(html, "id='pkg-version'")
	if !hasSyncIdentity {
		return false
	}

	// Fail if the only version assignment is clearly nested after health fetch success
	// with no prior sync paint. Heuristic: BUNDLE_VERSION or pkg-version write must
	// appear outside a sole fetch.then chain.
	jl := strings.ToLower(js)
	if strings.Contains(jl, "browser_agent_bundle_version") ||
		strings.Contains(jl, "pkg-version") {
		// Sync paint present in JS — good even if fetch also exists later.
		return true
	}
	// HTML-only static version element with bundle-sum is enough (MV3 external script).
	if strings.Contains(html, "bundle-sum.js") && hasVersionEl {
		return true
	}
	return hasVersionEl && hasSyncIdentity
}

// --- health timeout ---

// hasHealthFetchTimeout — health/status fetch uses AbortController or timeout abort.
// Bare fetch("/v1/health") without signal/abort does NOT satisfy.
func hasHealthFetchTimeout(js string) bool {
	if strings.TrimSpace(js) == "" {
		return false
	}
	jl := strings.ToLower(js)

	// Must touch health or control status path (or generic status refresh that fetches).
	touchesHealth := strings.Contains(jl, "/v1/health") ||
		strings.Contains(jl, "health") ||
		strings.Contains(jl, "ctrl-status") ||
		(strings.Contains(jl, "control") && strings.Contains(jl, "fetch"))
	if !touchesHealth && !strings.Contains(jl, "fetch(") {
		return false
	}

	// Strong signals: AbortController + abort/signal.
	hasAbortController := strings.Contains(js, "AbortController") ||
		strings.Contains(jl, "abortcontroller")
	hasSignal := strings.Contains(jl, "signal:") ||
		strings.Contains(jl, "signal =") ||
		strings.Contains(jl, "signal=") ||
		strings.Contains(jl, "{ signal") ||
		strings.Contains(jl, "signal ")
	hasAbortCall := strings.Contains(jl, ".abort(") ||
		strings.Contains(jl, "abort()")
	hasTimeout := strings.Contains(jl, "settimeout") ||
		strings.Contains(jl, "timeout") ||
		strings.Contains(jl, "deadline")

	if hasAbortController && (hasSignal || hasAbortCall) {
		// Prefer timeout pairing so abort is not open-ended.
		if hasTimeout || hasAbortCall {
			return true
		}
		// AbortController + signal alone is acceptable if used with fetch.
		if strings.Contains(jl, "fetch(") && hasSignal {
			return true
		}
	}

	// Promise.race with timeout rejecting health fetch.
	if strings.Contains(jl, "promise.race") && hasTimeout &&
		(strings.Contains(jl, "fetch(") || strings.Contains(jl, "health")) {
		return true
	}

	// AbortSignal.timeout (modern) or equivalent helper.
	if strings.Contains(jl, "abortsignal.timeout") ||
		strings.Contains(jl, "abortsignal.any") {
		return true
	}

	// Named helpers that implement timed health.
	helpers := []string{
		"fetchhealthwithtimeout",
		"healthwithtimeout",
		"checkhealthwithtimeout",
		"timedhealthfetch",
		"fetchwithtimeout",
		"healthfetch",
	}
	for _, h := range helpers {
		if !strings.Contains(jl, h) {
			continue
		}
		// Helper body or neighborhood must still abort/timeout.
		if hasAbortController || hasTimeout || hasSignal {
			return true
		}
		if body, ok := extractJSFunctionBody(js, h); ok {
			bl := strings.ToLower(body)
			if strings.Contains(bl, "abort") || strings.Contains(bl, "timeout") ||
				strings.Contains(bl, "signal") || strings.Contains(bl, "settimeout") {
				return true
			}
		}
	}
	return false
}

// --- progressive phases ---

// hasProgressiveStatusPhases — distinct UI hooks for daemon, session/ws, debugger.
// Single #ctrl-status "checking…" alone does NOT satisfy.
func hasProgressiveStatusPhases(html, js string) bool {
	combined := html + "\n" + js
	if strings.TrimSpace(combined) == "" {
		return false
	}
	cl := strings.ToLower(combined)

	// Count distinct phase surfaces. Accept ids, data-attrs, or getElementById names.
	daemon := phasePresent(cl, []string{
		"status-daemon", "status-control", "ctrl-status", "daemon-status",
		"health-status", "control-status", "phase-daemon", "phase-control",
		"phase-health", "data-phase=\"daemon\"", "data-phase='daemon'",
		"data-phase=\"control\"", "data-phase=\"health\"",
		"data-status-phase=\"daemon\"", "data-status-phase=\"control\"",
		"id=\"daemon\"", "id='daemon'",
	})
	session := phasePresent(cl, []string{
		"status-session", "status-ws", "session-status", "ws-status",
		"phase-session", "phase-ws", "websocket-status", "conn-status",
		"data-phase=\"session\"", "data-phase='session'",
		"data-phase=\"ws\"", "data-phase=\"websocket\"",
		"data-status-phase=\"session\"", "data-status-phase=\"ws\"",
		"id=\"session-status\"", "id='session-status'",
		"id=\"ws-status\"", "id='ws-status'",
	})
	debugger := phasePresent(cl, []string{
		"status-debugger", "status-attach", "status-armed", "debugger-status",
		"attach-status", "armed-status", "phase-debugger", "phase-attach",
		"phase-armed", "data-phase=\"debugger\"", "data-phase='debugger'",
		"data-phase=\"attach\"", "data-phase=\"armed\"",
		"data-status-phase=\"debugger\"", "data-status-phase=\"attach\"",
		"id=\"debugger-status\"", "id='debugger-status'",
		"id=\"attach-status\"", "id='attach-status'",
	})

	// Need all three dimensions. ctrl-status alone covers daemon only.
	if !(daemon && session && debugger) {
		// Also accept a single progressive container that lists all three labels
		// with structured rows (text hooks) AND three distinct element hooks.
		return false
	}

	// Reject pure single-status popup: only ctrl-status + conn-hint without
	// session/debugger hooks. (session && debugger already require extra hooks.)
	return true
}

func phasePresent(low string, markers []string) bool {
	for _, m := range markers {
		if strings.Contains(low, strings.ToLower(m)) {
			return true
		}
	}
	return false
}

// --- SW status API ---

// hasSWStatusAPI — background exposes status for popup: onMessage type and/or
// chrome.storage last-known write including session armed + debugger info.
// register-only onMessage does NOT satisfy. WS sendSessionStatus to control
// plane alone does NOT satisfy (that is server telemetry, not popup API).
func hasSWStatusAPI(bg string) bool {
	if strings.TrimSpace(bg) == "" {
		return false
	}
	bl := strings.ToLower(bg)

	// (1) onMessage status / getStatus / popupStatus handler.
	msgOK := false
	sec := listenerNeighborhood(bg, "onMessage", 2200)
	if sec == "" {
		sec = listenerNeighborhood(bg, "runtime.onMessage", 2200)
	}
	if sec != "" {
		sl := strings.ToLower(sec)
		// Must handle a status-like type (not only register).
		statusTypes := []string{
			`"status"`, `'status'`, `"getstatus"`, `'getstatus'`,
			`"get_status"`, `'get_status'`, `"popupstatus"`, `'popupstatus'`,
			`"popup_status"`, `'popup_status'`, `"getpopupstatus"`,
			`"lastknown"`, `'lastknown'`, `"last_known"`,
			"msg.type === \"status\"", "msg.type === 'status'",
			"msg.type==\"status\"", "msg.type=='status'",
			"type === \"status\"", "type === 'status'",
			"type === \"getstatus\"", "type === 'getstatus'",
			"type === \"popupstatus\"", "type === 'popupstatus'",
		}
		for _, st := range statusTypes {
			if strings.Contains(sl, strings.ToLower(st)) {
				msgOK = true
				break
			}
		}
		// Soft path still requires an explicit status *message type* string with quotes
		// (not changeInfo.status / pushStatusForConnectedSessions name alone).
		if !msgOK {
			explicitType := strings.Contains(sl, "\"status\"") || strings.Contains(sl, "'status'") ||
				strings.Contains(sl, "\"getstatus\"") || strings.Contains(sl, "'getstatus'") ||
				strings.Contains(sl, "\"get_status\"") || strings.Contains(sl, "'get_status'") ||
				strings.Contains(sl, "\"popupstatus\"") || strings.Contains(sl, "'popupstatus'") ||
				strings.Contains(sl, "getstatus") || strings.Contains(sl, "popupstatus")
			handlesMsg := strings.Contains(sl, "msg.type") || strings.Contains(sl, "type ===") ||
				strings.Contains(sl, "type==") || strings.Contains(sl, "case ")
			if explicitType && handlesMsg {
				msgOK = true
			}
		}
	}

	// Named status helpers that build popup payload.
	helperNames := []string{
		"getPopupStatus", "buildPopupStatus", "getStatusForPopup",
		"handleStatusMessage", "handleGetStatus", "popupStatus",
		"getExtensionStatus", "collectStatus", "buildStatusPayload",
		"getLastKnownStatus", "writeLastKnownStatus", "persistStatus",
		"publishStatus", "updateStoredStatus",
	}
	hasHelper := false
	for _, h := range helperNames {
		if strings.Contains(bg, h) || strings.Contains(bl, strings.ToLower(h)) {
			hasHelper = true
			break
		}
	}

	// (2) storage write of last-known with session + debugger/attach signals.
	storageWrite := (strings.Contains(bl, "storage.session.set") ||
		strings.Contains(bl, "storage.local.set") ||
		strings.Contains(bl, "chrome.storage.session") && strings.Contains(bl, ".set") ||
		strings.Contains(bl, "chrome.storage.local") && strings.Contains(bl, ".set") ||
		(strings.Contains(bl, "chrome.storage") && strings.Contains(bl, ".set(")))
	storageStatusful := storageWrite &&
		(strings.Contains(bl, "status") || strings.Contains(bl, "lastknown") ||
			strings.Contains(bl, "last_known") || strings.Contains(bl, "popup"))

	// Payload richness: session armed and debugger/attach.
	hasSessionInfo := strings.Contains(bl, "session") &&
		(strings.Contains(bl, "ws") || strings.Contains(bl, "websocket") ||
			strings.Contains(bl, "connected") || strings.Contains(bl, "armed") ||
			strings.Contains(bl, "sessions"))
	hasDebuggerInfo := strings.Contains(bl, "attachedtab") ||
		strings.Contains(bl, "attached_tab") ||
		strings.Contains(bl, "debugger") ||
		strings.Contains(bl, "attach") &&
			(strings.Contains(bl, "status") || strings.Contains(bl, "popup") ||
				strings.Contains(bl, "lastknown") || strings.Contains(bl, "armed"))

	// Message API with session+debugger payload is enough.
	if msgOK {
		// Prefer evidence of session/debugger in status helper or onMessage section.
		if hasSessionInfo || hasDebuggerInfo || hasHelper {
			return true
		}
		// Message type alone with sendResponse is weak but acceptable if status type is explicit.
		if sec != "" {
			sl := strings.ToLower(sec)
			if strings.Contains(sl, "sendresponse") || strings.Contains(sl, "return true") ||
				strings.Contains(sl, "sessions") || strings.Contains(sl, "attached") {
				return true
			}
		}
	}

	// Storage last-known write with session + debugger.
	if storageStatusful && hasSessionInfo && hasDebuggerInfo {
		return true
	}
	if storageStatusful && hasHelper && (hasSessionInfo || hasDebuggerInfo) {
		return true
	}

	// Helper that both builds status and is used with storage or onMessage.
	if hasHelper {
		for _, h := range helperNames {
			body, ok := extractJSFunctionBody(bg, h)
			if !ok {
				continue
			}
			bbl := strings.ToLower(body)
			if (strings.Contains(bbl, "session") || strings.Contains(bbl, "attached") ||
				strings.Contains(bbl, "debugger") || strings.Contains(bbl, "ws")) &&
				(msgOK || storageWrite || strings.Contains(bbl, "attached") ||
					strings.Contains(bbl, "session")) {
				return true
			}
		}
	}
	return false
}

// --- last-known paint ---

// hasLastKnownStatusPaint — popup reads chrome.storage session/local for last status
// before or in parallel with live health (not only after successful health).
func hasLastKnownStatusPaint(js string) bool {
	if strings.TrimSpace(js) == "" {
		return false
	}
	jl := strings.ToLower(js)

	readsStorage := strings.Contains(jl, "storage.session.get") ||
		strings.Contains(jl, "storage.local.get") ||
		strings.Contains(jl, "chrome.storage.session") ||
		strings.Contains(jl, "chrome.storage.local") ||
		(strings.Contains(jl, "chrome.storage") && strings.Contains(jl, ".get"))
	if !readsStorage {
		// sendMessage to SW for last-known / status also counts as last-known path
		// when paired with status type (SW is source of last-known).
		if strings.Contains(jl, "runtime.sendmessage") ||
			strings.Contains(jl, "chrome.runtime.sendmessage") {
			if strings.Contains(jl, "status") || strings.Contains(jl, "lastknown") ||
				strings.Contains(jl, "last_known") || strings.Contains(jl, "getstatus") ||
				strings.Contains(jl, "popupstatus") {
				// Still prefer storage; message-only is OK for "last-known" if SW caches.
				return true
			}
		}
		return false
	}

	// Should look status-related (not unrelated storage).
	statusish := strings.Contains(jl, "status") ||
		strings.Contains(jl, "lastknown") ||
		strings.Contains(jl, "last_known") ||
		strings.Contains(jl, "last-known") ||
		strings.Contains(jl, "popupstatus") ||
		strings.Contains(jl, "session") ||
		strings.Contains(jl, "armed") ||
		strings.Contains(jl, "debugger") ||
		strings.Contains(jl, "attached")
	if !statusish {
		// storage.get with progressive phase update still OK if it paints status els.
		statusish = strings.Contains(jl, "ctrl-status") ||
			strings.Contains(jl, "status-session") ||
			strings.Contains(jl, "status-debugger") ||
			strings.Contains(jl, "textcontent")
	}
	return statusish
}
```
