# Scenario

**Feature**: Session attach gate — attach only while a session control tab is open; detach on last leave

```
# static: read background.js attach/leave lifecycle
Test Client -> assert detach-on-leave, attach gate, multi-tab recount, reuse-while-open

# e2e: real browser attach gate behavior
RunDaemon -> POST /v1/sessions -> playwright-debug --extension
Session control tab open -> eval succeeds
Last control tab leaves -> subsequent eval fails (no sticky leftover attach)
Two control tabs, close one -> stays armed; eval still succeeds
```

## Preconditions

- **ModuleRoot** = workspace root (`filepath.Join(DOCTEST_ROOT, "..", "..")`).
- Classic TDD: intended attach-gate policy **not** implemented — sticky attach,
  no detach on `unregisterSession`, leave handling only on single `entry.tabId`.
  Ext-source leaves **RED** expected (except reuse regression may already pass).
- Ext-source leaves: no browser; read `Chrome-Ext-Browser-Agent/public/background.js`
  (also accept `build/` / `src/` fallbacks).
- E2e leaves: `playwright-debug` on PATH; Chromium for Playwright; skip when tool absent.
- Daemon phases 1–9 (`RunDaemon`, `POST /v1/sessions`, per-session extension WS).
- Scope: **Chrome Browser Agent extension only** (not browser-trace, not Firefox).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Grouping Setup sets `Mode` (`ext-source` or `e2e`).
3. E2e grouping allocates temp `BaseDir` and default timeouts.
4. Leaf Setup sets `ExtSourceTarget` or `PlaywrightOp` + `SessionID`.

## Context

- Spec version **0.0.2**.
- Complements `browser-agent-session-tab-targeting` (attach reuse / tab switch) and
  `browser-agent-active-tab-routing` (active tab pick) with **session-page attach gate**.
- Playwright uses `--headed` (MV3 extensions may not load in classic headless).
- Observable e2e strategy: job success/fail + extension-connected — **not** Chrome
  debugger infobar DOM scraping.

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
		// method-style: name( or name = function / name = async function
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

func containsDetachCall(low string) bool {
	return strings.Contains(low, "detachdebugger") ||
		strings.Contains(low, "chrome.debugger.detach") ||
		strings.Contains(low, "debugger.detach")
}

// hasDetachOnSessionLeave — last session-page leave / unregister tears down
// chrome.debugger for that session (not only WS close; not only detach-on-tab-switch).
// Current master: unregisterSession only closes WS — RED.
func hasDetachOnSessionLeave(text string) bool {
	// Pattern A: unregisterSession body invokes detach for the session.
	if body, ok := extractJSFunctionBody(text, "unregisterSession"); ok {
		bl := strings.ToLower(body)
		if containsDetachCall(bl) {
			return true
		}
		// Calls a named session-attach teardown helper from unregister.
		if strings.Contains(bl, "detach") &&
			(strings.Contains(bl, "session") || strings.Contains(bl, "attach")) {
			return true
		}
	}

	low := strings.ToLower(text)

	// Pattern B: explicit helper names for session detach on leave (wired by implementer).
	// Intentionally strict — do not treat tab-switch detach inside attachDebuggerForSession
	// or telemetry helpers as leave-detach.
	helpers := []string{
		"detachdebuggerforsession",
		"detachsessiondebugger",
		"releasedebuggerforsession",
		"teardownsessionattach",
		"clearsessionattach",
		"detachonsessionleave",
		"releasesessionattach",
		"detachsessionattach",
	}
	for _, h := range helpers {
		if strings.Contains(low, h) {
			return true
		}
	}

	// Pattern C: leave handlers (onRemoved / onUpdated navigate-away) call detach when
	// remaining control tabs hit 0 — inspect listener neighborhoods only (not whole file).
	for _, marker := range []string{"tabs.onRemoved", "tabs.onUpdated"} {
		sec := listenerNeighborhood(text, marker, 1200)
		sl := strings.ToLower(sec)
		if !containsDetachCall(sl) {
			continue
		}
		// Must couple detach to session attach state or remaining-control-tab logic —
		// not a bare detach elsewhere in a large neighborhood by accident.
		if (strings.Contains(sl, "sessionattachstate") || strings.Contains(sl, "attachedtabid") ||
			strings.Contains(sl, "detachdebuggerforsession") || strings.Contains(sl, "releasesession")) &&
			(strings.Contains(sl, "remaining") || strings.Contains(sl, "length") ||
				strings.Contains(sl, "count") || strings.Contains(sl, "unregister")) {
			return true
		}
	}
	return false
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

// hasAttachGateRequiresSessionPage — attach path refuses when no open control tab.
// Current master: attachDebuggerForSession always attaches — RED.
// Note: "session page not bound (windowId missing)" is routing, not the attach gate.
func hasAttachGateRequiresSessionPage(text string) bool {
	low := strings.ToLower(text)

	// Named gate helpers (preferred implementer surface).
	named := []string{
		"ensuresessionpageopen",
		"requireopensessionpage",
		"assertsessionpageopen",
		"hassessioncontroltab",
		"hasopensessionpage",
		"countopensessionpages",
		"countsessioncontroltabs",
		"sessionpagepresentforattach",
		"attachgaterequires",
		"guarddebuggerattach",
		"canattachforsession",
		"mayattachdebuggerforsession",
		"refuseattachwithoutsessionpage",
		"assertattachgate",
	}
	for _, n := range named {
		if strings.Contains(low, n) {
			return true
		}
	}

	// Gate check inside attach path functions (not pickTarget / telemetry).
	attachBody, _ := extractJSFunctionBody(text, "attachDebuggerForSession")
	withBody, _ := extractJSFunctionBody(text, "withDebuggerForSession")
	combined := strings.ToLower(attachBody + "\n" + withBody)
	if strings.TrimSpace(combined) == "" {
		return false
	}

	// Refuse copy for missing control tab (must appear in attach path, not only pickTarget).
	refusePhrases := []string{
		"no open session page",
		"no session page open",
		"session page is not open",
		"no open session control",
		"session control tab required",
		"session page required for attach",
		"attach refused",
		"cannot attach without",
		"no control tab open",
		"attach gate",
	}
	hasRefuseInAttachPath := false
	for _, p := range refusePhrases {
		if strings.Contains(combined, p) {
			hasRefuseInAttachPath = true
			break
		}
	}

	// Attach path must check open control tab before chrome.debugger.attach.
	checksOpenPage := strings.Contains(combined, "hasopensessionpage") ||
		strings.Contains(combined, "countopensession") ||
		strings.Contains(combined, "ensuresessionpage") ||
		strings.Contains(combined, "requireopensession") ||
		strings.Contains(combined, "sessionpageopen") ||
		strings.Contains(combined, "session control") ||
		(strings.Contains(combined, "tabs.query") &&
			(strings.Contains(combined, "go?session") || strings.Contains(combined, "issessiongopageurl") ||
				strings.Contains(combined, "/go")))

	refuses := hasRefuseInAttachPath ||
		((strings.Contains(combined, "throw") || strings.Contains(combined, "reject")) &&
			(strings.Contains(combined, "session page") || strings.Contains(combined, "control tab") ||
				strings.Contains(combined, "attach") && strings.Contains(combined, "open")))

	return checksOpenPage && refuses
}

// hasMultiSessionTabRecount — leave path re-queries open control tabs; not only entry.tabId.
// Current master: onRemoved/onUpdated only match entry.tabId — RED.
// Do not treat collectSessionPageTelemetry (hello telemetry) as leave recount.
func hasMultiSessionTabRecount(text string) bool {
	low := strings.ToLower(text)

	// Named recount helpers (leave / gate policy — not telemetry).
	helpers := []string{
		"countopensessionpages",
		"countsessioncontroltabs",
		"listsessioncontroltabs",
		"querysessioncontroltabs",
		"querysessionpagesforleave",
		"listsessionpagesforleave",
		"remainingsessionpages",
		"hasremainingsessionpage",
		"hasopensessionpage",
		"recountsessionpages",
		"findremainingsessionpages",
		"opensessionpagesforsession",
		"sessionpagesinwindow",
		"rebindsessiontabid",
	}
	for _, h := range helpers {
		if strings.Contains(low, h) {
			return true
		}
	}

	// Leave listeners must do more than `entry.tabId === closedTabId`.
	// Inspect onRemoved / onUpdated neighborhoods for recount (query + go filter),
	// not the whole file (avoids false GREEN from collectSessionPageTelemetry / pickTarget).
	for _, marker := range []string{"tabs.onRemoved", "tabs.onUpdated"} {
		sec := listenerNeighborhood(text, marker, 900)
		if sec == "" {
			continue
		}
		sl := strings.ToLower(sec)
		// Current master onRemoved: only entry.tabId === tabId → unregisterSession.
		// Desired: query remaining /go?session= tabs (or call a recount helper) before
		// unregister/detach decision.
		usesRecount := strings.Contains(sl, "countopensession") ||
			strings.Contains(sl, "remainingsession") ||
			strings.Contains(sl, "hasopensessionpage") ||
			strings.Contains(sl, "hasremainingsession") ||
			strings.Contains(sl, "recountsession") ||
			strings.Contains(sl, "rebindsession") ||
			strings.Contains(sl, "querysessioncontrol") ||
			strings.Contains(sl, "listsessioncontrol") ||
			(strings.Contains(sl, "tabs.query") &&
				(strings.Contains(sl, "issessiongopageurl") || strings.Contains(sl, "go?session") ||
					strings.Contains(sl, "parsegosessionfromurl")))
		if usesRecount {
			return true
		}
	}
	return false
}

// hasAttachReuseWhileSessionOpen — regression: reuse attach + detach on switch.
func hasAttachReuseWhileSessionOpen(text string) bool {
	low := strings.ToLower(text)
	hasReuse := strings.Contains(low, "attachedtabs.has") ||
		strings.Contains(low, "already attached") ||
		(strings.Contains(low, "attachedtabid") && strings.Contains(low, "=== tabid"))
	hasDetach := containsDetachCall(low)
	hasSwitch := (strings.Contains(low, "attachedtabid") &&
		(strings.Contains(low, "!==") || strings.Contains(low, "!="))) ||
		(strings.Contains(low, "tab_id") && (strings.Contains(low, "switch") || strings.Contains(low, "different")))
	hasSerialize := strings.Contains(low, "attachlock") ||
		strings.Contains(low, "attachmutex") ||
		strings.Contains(low, "attachqueue") ||
		strings.Contains(low, "serializ")
	return hasReuse && hasDetach && (hasSwitch || hasSerialize)
}
```
