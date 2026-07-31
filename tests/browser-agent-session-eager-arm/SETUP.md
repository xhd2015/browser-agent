# Scenario

**Feature**: Session eager arm — auto-attach capturable tabs when window is armed

```
# static: read background.js eager auto-attach call sites
Test Client -> assert auto-attach on /go register
Test Client -> assert auto-attach after create_tab
Test Client -> assert auto-attach on same-window navigate/update
Test Client -> assert skip non-capturable
Test Client -> assert other-window not targeted (entry.windowId scope)
```

## Preconditions

- **ModuleRoot** = workspace root (`filepath.Join(DOCTEST_ROOT, "..", "..")`).
- Classic TDD **P2**: eager auto-attach on register / create_tab / same-window
  navigate. Current job-only attach is **obsolete** for armed windows.
- Ext-source leaves: no browser; read `Chrome-Ext-Browser-Agent/public/background.js`
  (also accept `build/` / `src/` fallbacks).
- Scope: **Chrome Browser Agent extension only** (not browser-trace, not Firefox).
- Parallel-safe: pure FS reads; no `t.Setenv` / `os.Chdir` in harness.
- MECE with `browser-agent-session-attach-gate` (P1 set/gate/leave) — do not
  re-assert multi-attach keep peers or detach-all here.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Grouping Setup sets `Mode = ModeExtSource`.
3. Leaf Setup sets `ExtSourceTarget`.

## Context

- Spec version **0.0.2**.
- Complements attach-gate (when/whether attach is allowed + set shape) with
  **when attach is initiated** while the window is armed.
- Last `/go` leave still disarms (P1) — not re-tested here.
- Prefer `attachDebuggerForSession` so attach gate + multi-set stay intact.

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

func containsAttachForSessionCall(low string) bool {
	return strings.Contains(low, "attachdebuggerforsession") ||
		strings.Contains(low, "eagerattach") ||
		strings.Contains(low, "autoattach") ||
		(strings.Contains(low, "attachdebugger(") &&
			(strings.Contains(low, "session") || strings.Contains(low, "tabid")))
}

// hasNamedEagerAttachHelper — preferred implementer surface for P2.
func hasNamedEagerAttachHelper(text string) bool {
	low := strings.ToLower(text)
	helpers := []string{
		"eagerattachtab",
		"eagerattachforsession",
		"autoattachtab",
		"autoattachforsession",
		"maybeattachtabforsession",
		"maybeeagerattach",
		"attachifcapturable",
		"attachcapturabletab",
		"eagerarmattach",
		"armwindowattach",
		"autoattachonsessionarm",
		"maybeattachdebuggerforsession",
	}
	for _, h := range helpers {
		if strings.Contains(low, h) {
			return true
		}
	}
	return false
}

// bodyEagerAttachesTab — function body invokes session-scoped attach for a tabId.
func bodyEagerAttachesTab(body string) bool {
	if strings.TrimSpace(body) == "" {
		return false
	}
	bl := strings.ToLower(body)
	// Direct session attach (preferred).
	if strings.Contains(bl, "attachdebuggerforsession") {
		return true
	}
	// Named eager helper.
	if hasNamedEagerAttachHelper(body) {
		return true
	}
	// attachDebugger(tabId) only counts if clearly session-scoped context in body.
	if strings.Contains(bl, "attachdebugger(") &&
		(strings.Contains(bl, "sessionid") || strings.Contains(bl, "session_id") ||
			strings.Contains(bl, "sessionattach") || strings.Contains(bl, "attachedtabids")) {
		return true
	}
	return false
}

// hasRegisterConnectsWithoutDebuggerAttach — /go register binds + connectSession,
// and must NOT debugger-attach (popup UX: attach freezes default_popup).
func hasRegisterConnectsWithoutDebuggerAttach(text string) bool {
	body, ok := extractJSFunctionBody(text, "handleRegisterMessage")
	if !ok {
		return false
	}
	bl := strings.ToLower(body)
	if !strings.Contains(bl, "connectsession") {
		return false
	}
	// Register must not call session-scoped debugger attach helpers.
	forbidden := []string{
		"maybeeagerattach",
		"attachdebuggerforsession",
		"eagerattachtab",
		"autoattachtab",
	}
	for _, f := range forbidden {
		if strings.Contains(bl, f) {
			return false
		}
	}
	return true
}

// hasAutoAttachOnGoRegister — legacy name; now means WS-only register (popup UX).
func hasAutoAttachOnGoRegister(text string) bool {
	return hasRegisterConnectsWithoutDebuggerAttach(text)
}

// hasAutoAttachOnCreateTab — createTabInSession / handleCreateTabJob attaches new tab.
// chrome.tabs.create alone does NOT satisfy.
func hasAutoAttachOnCreateTab(text string) bool {
	bodies := []string{}
	for _, name := range []string{
		"createTabInSession",
		"handleCreateTabJob",
		"eagerAttachCreatedTab",
		"autoAttachCreatedTab",
	} {
		if body, ok := extractJSFunctionBody(text, name); ok {
			bodies = append(bodies, body)
		}
	}
	combined := strings.Join(bodies, "\n")
	if bodyEagerAttachesTab(combined) {
		// Prefer evidence create path actually creates then attaches.
		cl := strings.ToLower(combined)
		if strings.Contains(cl, "tabs.create") || strings.Contains(cl, "tab.id") ||
			strings.Contains(cl, "tab_id") || strings.Contains(cl, "created") ||
			strings.Contains(cl, "attachdebuggerforsession") {
			return true
		}
	}

	// createTabInSession must itself attach (or call helper that attaches).
	if body, ok := extractJSFunctionBody(text, "createTabInSession"); ok {
		bl := strings.ToLower(body)
		if strings.Contains(bl, "attachdebuggerforsession") {
			return true
		}
		// Calls a helper that is defined to attach.
		if (strings.Contains(bl, "eagerattach") || strings.Contains(bl, "autoattach") ||
			strings.Contains(bl, "maybeattach") || strings.Contains(bl, "attachifcapturable") ||
			strings.Contains(bl, "attachcapturable")) &&
			(hasNamedEagerAttachHelper(text) || strings.Contains(strings.ToLower(text), "attachdebuggerforsession")) {
			return true
		}
	}
	return false
}

// hasAutoAttachOnSameWindowNavigate — tabs.onUpdated / onCreated auto-attaches
// capturable tabs in the armed session window.
// maybeRegisterGoTab alone (register only) does NOT satisfy.
func hasAutoAttachOnSameWindowNavigate(text string) bool {
	// Dedicated helper wired from listeners is enough if listener neighborhood
	// references attach for non-control tabs.
	for _, marker := range []string{"tabs.onUpdated", "tabs.onCreated"} {
		sec := listenerNeighborhood(text, marker, 1600)
		if sec == "" {
			continue
		}
		sl := strings.ToLower(sec)
		// Must attach (session-scoped), not only register /go.
		if !containsAttachForSessionCall(sl) &&
			!strings.Contains(sl, "eagerattach") &&
			!strings.Contains(sl, "autoattach") &&
			!strings.Contains(sl, "maybeattach") &&
			!strings.Contains(sl, "attachifcapturable") &&
			!strings.Contains(sl, "attachcapturable") {
			continue
		}
		// Should scope to session window and/or capturable.
		windowScoped := strings.Contains(sl, "windowid") ||
			strings.Contains(sl, "entry.window") ||
			strings.Contains(sl, "session") ||
			strings.Contains(sl, "armed")
		capturableAware := strings.Contains(sl, "iscapturable") ||
			strings.Contains(sl, "capturable") ||
			strings.Contains(sl, "attachdebuggerforsession") ||
			strings.Contains(sl, "eagerattach") ||
			strings.Contains(sl, "autoattach")
		if windowScoped && capturableAware {
			return true
		}
		// attachDebuggerForSession in onUpdated/onCreated neighborhood is strong signal.
		if strings.Contains(sl, "attachdebuggerforsession") {
			return true
		}
	}

	// Named navigate/create eager helpers present + used in file.
	if hasNamedEagerAttachHelper(text) {
		// Require listener wiring beyond /go register.
		for _, marker := range []string{"tabs.onUpdated", "tabs.onCreated"} {
			sec := listenerNeighborhood(text, marker, 1400)
			sl := strings.ToLower(sec)
			if strings.Contains(sl, "eagerattach") || strings.Contains(sl, "autoattach") ||
				strings.Contains(sl, "maybeattach") || strings.Contains(sl, "attachifcapturable") ||
				strings.Contains(sl, "attachcapturable") || strings.Contains(sl, "attachdebuggerforsession") {
				// Exclude pure maybeRegisterGoTab-only neighborhood (no attach words above would fail).
				if !strings.Contains(sl, "attach") {
					continue
				}
				// Reject if the only attach-like word is absent and only register exists.
				if strings.Contains(sl, "attachdebuggerforsession") ||
					strings.Contains(sl, "eagerattach") ||
					strings.Contains(sl, "autoattach") ||
					strings.Contains(sl, "maybeattachtab") ||
					strings.Contains(sl, "attachifcapturable") {
					return true
				}
			}
		}
	}
	return false
}

// hasEagerAttachSkipNonCapturable — eager auto-attach paths guard with isCapturableTabURL.
// Job-path isCapturableTabURL alone (pickTarget) does NOT satisfy without eager call sites.
func hasEagerAttachSkipNonCapturable(text string) bool {
	// Must have at least one eager trigger implemented.
	if !hasAutoAttachOnCreateTab(text) &&
		!hasAutoAttachOnSameWindowNavigate(text) &&
		!hasAutoAttachOnGoRegister(text) &&
		!hasNamedEagerAttachHelper(text) {
		return false
	}

	// isCapturableTabURL (or equivalent) must exist.
	hasCapturableFn := strings.Contains(text, "isCapturableTabURL") ||
		strings.Contains(strings.ToLower(text), "iscapturabletaburl") ||
		strings.Contains(strings.ToLower(text), "iscapturable")
	if !hasCapturableFn {
		return false
	}

	// Guard must appear in eager path bodies / helpers, not only pickTargetTabIdForSession.
	// Inspect createTabInSession, named eager helpers, and listener neighborhoods.
	sections := []string{}
	for _, name := range []string{
		"createTabInSession",
		"handleCreateTabJob",
		"handleRegisterMessage",
		"maybeRegisterGoTab",
		"eagerAttachTab",
		"eagerAttachForSession",
		"autoAttachTab",
		"autoAttachForSession",
		"maybeAttachTabForSession",
		"maybeEagerAttach",
		"attachIfCapturable",
		"attachCapturableTab",
	} {
		if body, ok := extractJSFunctionBody(text, name); ok {
			sections = append(sections, body)
		}
	}
	for _, marker := range []string{"tabs.onUpdated", "tabs.onCreated"} {
		if sec := listenerNeighborhood(text, marker, 1600); sec != "" {
			sections = append(sections, sec)
		}
	}
	combined := strings.ToLower(strings.Join(sections, "\n"))
	if strings.TrimSpace(combined) == "" {
		return false
	}

	// Eager section must both attach and check capturable (or chrome:// skip).
	attaches := containsAttachForSessionCall(combined) ||
		strings.Contains(combined, "eagerattach") ||
		strings.Contains(combined, "autoattach") ||
		strings.Contains(combined, "attachdebuggerforsession")
	guards := strings.Contains(combined, "iscapturable") ||
		strings.Contains(combined, "capturable") ||
		(strings.Contains(combined, "chrome://") &&
			(strings.Contains(combined, "skip") || strings.Contains(combined, "return") ||
				strings.Contains(combined, "!") || strings.Contains(combined, "not")))
	return attaches && guards
}

// hasEagerAttachScopedToSessionWindow — create + eager attach limited to entry.windowId.
// Other windows must not be auto-attached for the session.
func hasEagerAttachScopedToSessionWindow(text string) bool {
	// (1) createTabInSession must scope chrome.tabs.create to entry.windowId (regression).
	createBody, createOK := extractJSFunctionBody(text, "createTabInSession")
	if !createOK {
		return false
	}
	cl := strings.ToLower(createBody)
	createScoped := strings.Contains(cl, "windowid") &&
		(strings.Contains(cl, "entry.windowid") || strings.Contains(cl, "entry.window") ||
			(strings.Contains(cl, "entry") && strings.Contains(cl, "windowid"))) &&
		strings.Contains(cl, "tabs.create")
	if !createScoped {
		return false
	}

	// (2) At least one eager attach path must exist (P2) — window-scoped create alone
	// is create-tab product surface, not eager arm.
	if !hasAutoAttachOnCreateTab(text) && !hasAutoAttachOnSameWindowNavigate(text) {
		return false
	}

	// (3) Eager attach paths must reference session window scope.
	sections := []string{createBody}
	if body, ok := extractJSFunctionBody(text, "handleCreateTabJob"); ok {
		sections = append(sections, body)
	}
	for _, name := range []string{
		"eagerAttachTab", "eagerAttachForSession", "autoAttachTab",
		"autoAttachForSession", "maybeAttachTabForSession", "maybeEagerAttach",
		"attachIfCapturable", "attachCapturableTab",
	} {
		if body, ok := extractJSFunctionBody(text, name); ok {
			sections = append(sections, body)
		}
	}
	for _, marker := range []string{"tabs.onUpdated", "tabs.onCreated"} {
		if sec := listenerNeighborhood(text, marker, 1600); sec != "" {
			sections = append(sections, sec)
		}
	}
	combined := strings.ToLower(strings.Join(sections, "\n"))

	// Window isolation signals in eager sections.
	windowGuard := strings.Contains(combined, "entry.windowid") ||
		strings.Contains(combined, "windowid:") || // create props
		(strings.Contains(combined, "windowid") &&
			(strings.Contains(combined, "entry") || strings.Contains(combined, "session"))) ||
		(strings.Contains(combined, "tab.windowid") && strings.Contains(combined, "entry"))

	// Must not look like global auto-attach of all windows without session scope.
	// If onUpdated attaches without any windowId/session mention → fail.
	for _, marker := range []string{"tabs.onUpdated", "tabs.onCreated"} {
		sec := listenerNeighborhood(text, marker, 1600)
		if sec == "" {
			continue
		}
		sl := strings.ToLower(sec)
		if containsAttachForSessionCall(sl) || strings.Contains(sl, "attachdebuggerforsession") ||
			strings.Contains(sl, "eagerattach") || strings.Contains(sl, "autoattach") {
			if !strings.Contains(sl, "windowid") && !strings.Contains(sl, "session") {
				return false
			}
		}
	}

	return windowGuard
}
```
