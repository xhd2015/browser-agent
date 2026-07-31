# Scenario

**Feature**: Session attach gate + multi-tab attach set (policy B)

```
# static: read background.js attach/leave lifecycle (policy B)
Test Client -> assert multi-attach keep peers, same-tab reuse+lock,
               detach-all on leave, attach gate, multi-tab recount

# e2e: real browser attach gate behavior
RunDaemon -> POST /v1/sessions -> playwright-debug --extension
Session control tab open -> eval succeeds
Last control tab leaves -> subsequent eval fails (no sticky leftover attach)
Two control tabs, close one -> stays armed; eval still succeeds
```

## Preconditions

- **ModuleRoot** = workspace root (`filepath.Join(DOCTEST_ROOT, "..", "..")`).
- Classic TDD **policy B**: multi-tab attach set — attach B keeps A; leave detaches
  **all** tabs in the set. Current sticky single `attachedTabId` + switch-detach is
  **obsolete**. Multi-attach / detach-all leaves **RED** under sticky code.
- Ext-source leaves: no browser; read `Chrome-Ext-Browser-Agent/public/background.js`
  (also accept `build/` / `src/` fallbacks).
- E2e leaves: `playwright-debug` on PATH; Chromium for Playwright; skip when tool absent.
- Daemon phases 1–9 (`RunDaemon`, `POST /v1/sessions`, per-session extension WS).
- Scope: **Chrome Browser Agent extension only** (not browser-trace, not Firefox).
- Parallel-safe: pure FS reads; no `t.Setenv` / `os.Chdir` in harness.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Grouping Setup sets `Mode` (`ext-source` or `e2e`).
3. E2e grouping allocates temp `BaseDir` and default timeouts.
4. Leaf Setup sets `ExtSourceTarget` or `PlaywrightOp` + `SessionID`.

## Context

- Spec version **0.0.2**.
- Complements `browser-agent-session-tab-targeting` (tab_id / tab_index targeting)
  and `browser-agent-active-tab-routing` (active tab pick) with **session-page attach
  gate** + **multi-tab attach set**.
- Switch-detach (policy A) must not remain GREEN anywhere in this tree.
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

// hasLeaveDetachWiring — unregister / leave path invokes session debugger teardown.
// Does not alone prove multi-set detach-all (see hasDetachAllOnSessionLeave).
func hasLeaveDetachWiring(text string) bool {
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

	// Pattern B: explicit helper names for session detach on leave.
	helpers := []string{
		"detachdebuggerforsession",
		"detachsessiondebugger",
		"releasedebuggerforsession",
		"teardownsessionattach",
		"clearsessionattach",
		"detachonsessionleave",
		"releasesessionattach",
		"detachsessionattach",
		"detachallsessiontabs",
		"detachallattachedtabs",
	}
	for _, h := range helpers {
		if strings.Contains(low, h) {
			return true
		}
	}

	// Pattern C: leave listeners call detach when remaining control tabs hit 0.
	for _, marker := range []string{"tabs.onRemoved", "tabs.onUpdated"} {
		sec := listenerNeighborhood(text, marker, 1200)
		sl := strings.ToLower(sec)
		if !containsDetachCall(sl) {
			continue
		}
		if (strings.Contains(sl, "sessionattachstate") || strings.Contains(sl, "attachedtabid") ||
			strings.Contains(sl, "detachdebuggerforsession") || strings.Contains(sl, "releasesession") ||
			strings.Contains(sl, "detachsession")) &&
			(strings.Contains(sl, "remaining") || strings.Contains(sl, "length") ||
				strings.Contains(sl, "count") || strings.Contains(sl, "unregister")) {
			return true
		}
	}
	return false
}

// detachesAllFromSessionAttachSet — body detaches every tab in a multi-id collection.
// Single sticky `attachedTabId` one-shot detach does NOT satisfy policy B.
func detachesAllFromSessionAttachSet(body string) bool {
	if strings.TrimSpace(body) == "" {
		return false
	}
	bl := strings.ToLower(body)
	if !containsDetachCall(bl) {
		return false
	}

	// Multi-id collection markers (session attach set, not global attachedTabs alone).
	hasMultiField := strings.Contains(bl, "attachedtabids") ||
		strings.Contains(bl, "attachedtabset") ||
		strings.Contains(bl, "sessionattachedtabs") ||
		strings.Contains(bl, "sessionattachedids") ||
		strings.Contains(bl, "attachedids")

	// Iteration over a collection while detaching (for / forEach / for...of / Array.from).
	hasForLoop := (strings.Contains(bl, "for (") || strings.Contains(bl, "for(")) &&
		(strings.Contains(bl, "tab") || strings.Contains(bl, "id"))
	hasIterate := strings.Contains(bl, "foreach") ||
		strings.Contains(bl, "array.from") ||
		strings.Contains(bl, " of ") ||
		hasForLoop

	// Clear-set after detaching all members.
	hasClearSet := (strings.Contains(bl, ".clear(") || strings.Contains(bl, ".clear()")) &&
		(hasMultiField || strings.Contains(bl, "set") || strings.Contains(bl, "attach"))

	if hasMultiField && (hasIterate || hasClearSet || containsDetachCall(bl)) {
		// Reject pure sticky single: only attachedTabId (singular) without multi field / loop.
		return true
	}
	if hasIterate && containsDetachCall(bl) &&
		(hasMultiField || strings.Contains(bl, "new set") || strings.Contains(bl, "set(") ||
			strings.Contains(bl, "attached")) {
		// Still reject classic sticky body: single attachedTabId assign + one detach, no loop.
		if strings.Contains(bl, "attachedtabid") && !hasMultiField &&
			!strings.Contains(bl, "foreach") && !strings.Contains(bl, "array.from") &&
			!strings.Contains(bl, " of ") {
			return false
		}
		return true
	}
	return false
}

// hasDetachAllOnSessionLeave — leave/unregister tears down chrome.debugger for
// **every** tab in the session attach set (policy B). Sticky single-id detach fails.
func hasDetachAllOnSessionLeave(text string) bool {
	if !hasLeaveDetachWiring(text) {
		return false
	}

	// Prefer dedicated teardown helpers that iterate the attach set.
	for _, name := range []string{
		"detachSessionDebugger",
		"detachDebuggerForSession",
		"detachSessionAttach",
		"releaseSessionAttach",
		"teardownSessionAttach",
		"clearSessionAttach",
		"detachOnSessionLeave",
		"detachAllSessionTabs",
		"detachAllAttachedTabs",
		"releaseSessionDebugger",
	} {
		if body, ok := extractJSFunctionBody(text, name); ok {
			if detachesAllFromSessionAttachSet(body) {
				return true
			}
		}
	}

	// unregisterSession may inline multi-detach.
	if body, ok := extractJSFunctionBody(text, "unregisterSession"); ok {
		if detachesAllFromSessionAttachSet(body) {
			return true
		}
	}

	// handleSessionControlLeave may detach-all when remaining == 0.
	if body, ok := extractJSFunctionBody(text, "handleSessionControlLeave"); ok {
		if detachesAllFromSessionAttachSet(body) {
			return true
		}
	}

	return false
}

// hasAttachGateRequiresSessionPage — attach path refuses when no open control tab.
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
	for _, marker := range []string{"tabs.onRemoved", "tabs.onUpdated"} {
		sec := listenerNeighborhood(text, marker, 900)
		if sec == "" {
			continue
		}
		sl := strings.ToLower(sec)
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

// hasSessionAttachSetRepresentation — per-session attach state is a multi-tab set
// (not only a singular sticky attachedTabId).
func hasSessionAttachSetRepresentation(text string) bool {
	// Explicit multi-id field names (case-sensitive first for JS identifiers).
	for _, m := range []string{
		"attachedTabIds",
		"attachedTabSet",
		"sessionAttachedTabs",
		"sessionAttachedIds",
		"attachedIds",
	} {
		if strings.Contains(text, m) {
			return true
		}
	}
	low := strings.ToLower(text)
	for _, m := range []string{
		"attachedtabids",
		"attachedtabset",
		"sessionattachedtabs",
		"sessionattachedids",
	} {
		if strings.Contains(low, m) {
			return true
		}
	}

	// sessionAttachState typed / initialized with a Set of tab ids.
	// Inspect neighborhood of sessionAttachState declarations (not global attachedTabs Map alone).
	idx := 0
	for {
		i := strings.Index(text[idx:], "sessionAttachState")
		if i < 0 {
			break
		}
		abs := idx + i
		start := abs - 200
		if start < 0 {
			start = 0
		}
		end := abs + 350
		if end > len(text) {
			end = len(text)
		}
		sec := text[start:end]
		sl := strings.ToLower(sec)
		// Set<number> / new Set / attachedTabIds in state shape
		if (strings.Contains(sec, "Set") || strings.Contains(sl, "new set")) &&
			(strings.Contains(sl, "tab") || strings.Contains(sl, "number") ||
				strings.Contains(sl, "attach")) {
			// Avoid matching only Map<string, ... attachedTabId singular>
			if strings.Contains(sl, "attachedtabids") || strings.Contains(sl, "set<number") ||
				strings.Contains(sl, "set <number") || strings.Contains(sec, "Set<number>") ||
				strings.Contains(sec, "new Set") || strings.Contains(sec, "new Set(") {
				return true
			}
		}
		idx = abs + len("sessionAttachState")
	}
	return false
}

// hasSwitchDetachAntiPattern — attachDebuggerForSession detaches previous session
// tab solely because a different tabId is being attached (policy A sticky switch).
func hasSwitchDetachAntiPattern(attachBody string) bool {
	if strings.TrimSpace(attachBody) == "" {
		return false
	}
	// Precise sticky comparisons (current master pattern).
	switchMarkers := []string{
		"attachedTabId !== tabId",
		"attachedTabId != tabId",
		"attachedTabId!==tabId",
		"attachedTabId!=tabId",
		"state.attachedTabId !== tabId",
		"state.attachedTabId != tabId",
		"state.attachedTabId!==tabId",
		"state.attachedTabId!=tabId",
	}
	for _, m := range switchMarkers {
		if strings.Contains(attachBody, m) {
			return containsDetachCall(strings.ToLower(attachBody))
		}
	}
	bl := strings.ToLower(attachBody)
	// "detach previous" / "on tab switch" phrasing with a detach call.
	if containsDetachCall(bl) &&
		((strings.Contains(bl, "previous") && strings.Contains(bl, "detach")) ||
			(strings.Contains(bl, "tab switch") || strings.Contains(bl, "switch tab") ||
				strings.Contains(bl, "switching tab"))) {
		return true
	}
	return false
}

// attachesIntoSessionSet — attach path records tabId into a multi-tab set.
func attachesIntoSessionSet(attachBody string) bool {
	if strings.TrimSpace(attachBody) == "" {
		return false
	}
	bl := strings.ToLower(attachBody)
	// set.add(tabId) / attachedTabIds.add / push into multi collection
	if strings.Contains(bl, "attachedtabids") &&
		(strings.Contains(bl, ".add(") || strings.Contains(bl, "tabid")) {
		return true
	}
	if strings.Contains(bl, "attachedtabset") {
		return true
	}
	if (strings.Contains(bl, ".add(") || strings.Contains(bl, ".push(")) &&
		strings.Contains(bl, "tabid") &&
		(strings.Contains(bl, "attach") || strings.Contains(bl, "state.") ||
			strings.Contains(bl, "set")) {
		return true
	}
	return false
}

// hasMultiAttachKeepPeers — policy B: attach path keeps peer tabs attached.
// RED under sticky single attachedTabId + switch-detach.
func hasMultiAttachKeepPeers(text string) bool {
	attachBody, ok := extractJSFunctionBody(text, "attachDebuggerForSession")
	if !ok {
		return false
	}
	if hasSwitchDetachAntiPattern(attachBody) {
		return false
	}
	if !hasSessionAttachSetRepresentation(text) {
		return false
	}
	if !attachesIntoSessionSet(attachBody) {
		return false
	}
	return true
}

// hasSameTabReuseAndLock — reuse attach for same tabId + serialize per session.
// Does **not** require switch-detach (policy A obsolete).
func hasSameTabReuseAndLock(text string) bool {
	low := strings.ToLower(text)
	hasReuse := strings.Contains(low, "attachedtabs.has") ||
		strings.Contains(low, "already attached") ||
		(strings.Contains(low, "attachedtabids") && strings.Contains(low, ".has(")) ||
		(strings.Contains(low, "attachedtabset") && strings.Contains(low, ".has(")) ||
		(strings.Contains(low, "attachedtabid") && strings.Contains(low, "=== tabid"))
	hasSerialize := strings.Contains(low, "attachlock") ||
		strings.Contains(low, "attachmutex") ||
		strings.Contains(low, "attachqueue") ||
		strings.Contains(low, "serializ")
	return hasReuse && hasSerialize
}
```
