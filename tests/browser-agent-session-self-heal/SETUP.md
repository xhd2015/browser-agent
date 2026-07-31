# Scenario

**Feature**: Session self-heal — rediscover `/go` tabs on SW boot

```
# static: read background.js boot heal call sites
Test Client -> assert boot rediscover via tabs.query + /go?session=
Test Client -> assert reconnect WS for discovered sessions
Test Client -> assert re-attach (eager) on heal
Test Client -> assert empty sessions.keys() loop alone is insufficient
```

## Preconditions

- **ModuleRoot** = workspace root (`filepath.Join(DOCTEST_ROOT, "..", "..")`).
- Classic TDD **P3**: self-heal after reload / SW restart. Current
  `onInstalled` / `onStartup` only iterate empty `sessions.keys()` — **obsolete**.
- Ext-source leaves: no browser; read `Chrome-Ext-Browser-Agent/public/background.js`
  (also accept `build/` / `src/` fallbacks).
- Scope: **Chrome Browser Agent extension only** (not browser-trace, not Firefox).
- Parallel-safe: pure FS reads; no `t.Setenv` / `os.Chdir` in harness.
- MECE with:
  - `browser-agent-session-attach-gate` (P1 set/gate/leave)
  - `browser-agent-session-eager-arm` (P2 when attach is initiated while SW alive)
  — do not re-assert multi-attach or per-event eager triggers here.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Grouping Setup sets `Mode = ModeExtSource`.
3. Leaf Setup sets `ExtSourceTarget`.

## Context

- Spec version **0.0.1**.
- Complements eager-arm (attach while SW alive) with **boot rediscovery** when
  in-memory state is gone.
- Prefer reusing `parseGoSessionFromURL` / `maybeRegisterGoTab` /
  `handleRegisterMessage` / `maybeEagerAttach` / `attachDebuggerForSession`.
- Optional `chrome.storage` is fine but not required.

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

// selfHealHelperNames — preferred implementer surfaces for P3 boot heal.
func selfHealHelperNames() []string {
	return []string{
		"healSessions",
		"selfHeal",
		"selfHealSessions",
		"healOnBoot",
		"selfHealOnBoot",
		"rediscoverSessions",
		"rediscoverGoTabs",
		"rediscoverOpenGoTabs",
		"healOpenGoTabs",
		"healSessionsFromTabs",
		"restoreSessionsFromTabs",
		"restoreOpenSessions",
		"bootstrapSessions",
		"recoverSessions",
		"rehydrateSessions",
		"discoverOpenSessions",
		"rediscoverAndHealSessions",
		"healFromOpenGoTabs",
	}
}

// hasNamedSelfHealHelper — file declares a named P3 heal helper.
func hasNamedSelfHealHelper(text string) bool {
	low := strings.ToLower(text)
	for _, h := range selfHealHelperNames() {
		if strings.Contains(low, strings.ToLower(h)) {
			// Prefer function declaration signal when possible.
			if strings.Contains(text, "function "+h) ||
				strings.Contains(text, "function "+strings.ToLower(h)) ||
				strings.Contains(low, "function "+strings.ToLower(h)) ||
				strings.Contains(low, strings.ToLower(h)+" =") ||
				strings.Contains(low, strings.ToLower(h)+"(") {
				return true
			}
		}
	}
	return false
}

// extractSelfHealSections collects heal helper bodies + boot listener neighborhoods.
// Excludes pure job/telemetry tabs.query call sites by preferring heal-named helpers
// and onInstalled/onStartup neighborhoods.
func extractSelfHealSections(text string) string {
	var parts []string
	for _, name := range selfHealHelperNames() {
		if body, ok := extractJSFunctionBody(text, name); ok {
			parts = append(parts, body)
		}
	}
	// Also accept camelCase variants that may be async function declarations
	// already covered by extractJSFunctionBody via "async function name(".
	for _, marker := range []string{
		"chrome.runtime.onInstalled",
		"runtime.onInstalled",
		"chrome.runtime.onStartup",
		"runtime.onStartup",
	} {
		if sec := listenerNeighborhood(text, marker, 1800); sec != "" {
			parts = append(parts, sec)
		}
	}
	return strings.Join(parts, "\n")
}

// sectionQueriesTabsForGo — section uses tabs.query and looks for /go session pages.
func sectionQueriesTabsForGo(section string) bool {
	if strings.TrimSpace(section) == "" {
		return false
	}
	sl := strings.ToLower(section)
	hasQuery := strings.Contains(sl, "tabs.query") ||
		strings.Contains(sl, "querysessioncontroltabs") ||
		strings.Contains(sl, "chrome.tabs.query")
	if !hasQuery {
		return false
	}
	// Must parse or match /go session control URLs — not a generic window tab list.
	parsesGo := strings.Contains(sl, "parsegosessionfromurl") ||
		strings.Contains(sl, "issessiongopageurl") ||
		strings.Contains(sl, "/go?session=") ||
		(strings.Contains(sl, "/go") &&
			(strings.Contains(sl, "session") || strings.Contains(sl, "parsesession") ||
				strings.Contains(sl, "sessionid")))
	return parsesGo
}

// sectionReconnectsDiscovered — section reconnects sessions found via tab rediscover,
// not only via empty sessions.keys() → connectSession.
func sectionReconnectsDiscovered(section string) bool {
	if strings.TrimSpace(section) == "" {
		return false
	}
	if !sectionQueriesTabsForGo(section) {
		return false
	}
	sl := strings.ToLower(section)

	// Strong: register path binds discovered control tabs then connects.
	if strings.Contains(sl, "handleregistermessage") ||
		strings.Contains(sl, "mayberegistergotab") {
		return true
	}

	// Explicit bind of discovered session + connectSession.
	// Require evidence that session id comes from parsed tab URL / entry create,
	// not solely from sessions.keys().
	hasConnect := strings.Contains(sl, "connectsession")
	if !hasConnect {
		return false
	}
	bindsDiscovered := strings.Contains(sl, "getorcreatesessionentry") ||
		strings.Contains(sl, "parsegosessionfromurl") ||
		strings.Contains(sl, "parsed.session") ||
		strings.Contains(sl, "parsed.sessionid") ||
		(strings.Contains(sl, "sessionid") &&
			(strings.Contains(sl, "tab.id") || strings.Contains(sl, "tabid") ||
				strings.Contains(sl, "t.id") || strings.Contains(sl, "windowid")))

	if !bindsDiscovered {
		return false
	}

	// If the only connectSession is clearly keys-only and there is no register path,
	// reject even when query+parse co-exist in the same neighborhood.
	// Heuristic: keys loop present without bind-from-tab signals beyond keys → fail.
	// bindsDiscovered already requires tab-derived bind; keys loop may coexist.
	return true
}

// sectionOnlyEmptySessionsKeysLoop — classic obsolete boot: only sessions.keys → connectSession.
// True when the section reconnects known map keys but does NOT query tabs for /go rediscover.
func sectionOnlyEmptySessionsKeysLoop(section string) bool {
	if strings.TrimSpace(section) == "" {
		return true
	}
	sl := strings.ToLower(section)
	hasKeysLoop := strings.Contains(sl, "sessions.keys") ||
		(strings.Contains(sl, "sessions") && strings.Contains(sl, ".keys("))
	hasConnect := strings.Contains(sl, "connectsession")
	hasRediscover := sectionQueriesTabsForGo(section) ||
		strings.Contains(sl, "parsegosessionfromurl") ||
		strings.Contains(sl, "mayberegistergotab") ||
		strings.Contains(sl, "handleregistermessage")
	// Obsolete if it loops keys+connect without rediscover, OR has neither heal nor rediscover.
	if hasKeysLoop && hasConnect && !hasRediscover {
		return true
	}
	if !hasRediscover && !hasNamedHealCall(sl) {
		// Neighborhood that only logs / empty body counts as insufficient.
		return true
	}
	// Query present but reconnect is still only keys-loop (no discovered bind) → obsolete-ish.
	if hasRediscover && hasKeysLoop && hasConnect && !sectionReconnectsDiscovered(section) {
		return true
	}
	return false
}

func hasNamedHealCall(low string) bool {
	for _, h := range selfHealHelperNames() {
		if strings.Contains(low, strings.ToLower(h)) {
			return true
		}
	}
	return false
}

// hasBootRediscoverGoTabs — onInstalled/onStartup/init queries tabs and parses /go?session=.
// Empty sessions.keys() loop alone does NOT satisfy.
func hasBootRediscoverGoTabs(text string) bool {
	// (1) Named heal helper body queries tabs for /go and is wired from boot (or is entrypoint).
	for _, name := range selfHealHelperNames() {
		if body, ok := extractJSFunctionBody(text, name); ok {
			if sectionQueriesTabsForGo(body) {
				if healHelperWiredFromBoot(text, name) {
					return true
				}
				// Entrypoint helper called from boot neighborhood is preferred; bare declaration
				// without boot wiring does not count (avoid orphan helpers greening boot).
			}
		}
	}

	// (2) onInstalled / onStartup neighborhood itself queries + parses /go.
	for _, marker := range []string{
		"chrome.runtime.onInstalled",
		"runtime.onInstalled",
		"chrome.runtime.onStartup",
		"runtime.onStartup",
	} {
		sec := listenerNeighborhood(text, marker, 1800)
		if sectionQueriesTabsForGo(sec) {
			return true
		}
		// Listener delegates to a heal helper that rediscovers.
		sl := strings.ToLower(sec)
		if hasNamedHealCall(sl) {
			for _, name := range selfHealHelperNames() {
				if !strings.Contains(sl, strings.ToLower(name)) {
					continue
				}
				if body, ok := extractJSFunctionBody(text, name); ok && sectionQueriesTabsForGo(body) {
					return true
				}
			}
		}
	}

	// (3) Top-level init: rediscover helper is defined and invoked (decl + call).
	// Covers SW boot without relying solely on onInstalled listener wiring.
	low := strings.ToLower(text)
	for _, name := range selfHealHelperNames() {
		nl := strings.ToLower(name)
		body, ok := extractJSFunctionBody(text, name)
		if !ok || !sectionQueriesTabsForGo(body) {
			continue
		}
		if !looksLikeBootHealEntrypoint(name) {
			continue
		}
		// function healSessions(… is one "healSessions("; healSessions() call is a second.
		if strings.Count(low, nl+"(") >= 2 {
			return true
		}
	}

	return false
}

func looksLikeBootHealEntrypoint(name string) bool {
	low := strings.ToLower(name)
	markers := []string{"heal", "rediscover", "restore", "bootstrap", "recover", "rehydrate", "discover"}
	for _, m := range markers {
		if strings.Contains(low, m) {
			return true
		}
	}
	return false
}

func healHelperWiredFromBoot(text, name string) bool {
	nl := strings.ToLower(name)
	for _, marker := range []string{
		"chrome.runtime.onInstalled",
		"runtime.onInstalled",
		"chrome.runtime.onStartup",
		"runtime.onStartup",
	} {
		sec := listenerNeighborhood(text, marker, 1800)
		if strings.Contains(strings.ToLower(sec), nl) {
			return true
		}
	}
	return false
}

// hasReconnectWSOnHeal — heal path calls connectSession / register for discovered sessions.
// sessions.keys() → connectSession alone does NOT satisfy.
func hasReconnectWSOnHeal(text string) bool {
	// Named heal helper: rediscover + reconnect discovered.
	for _, name := range selfHealHelperNames() {
		body, ok := extractJSFunctionBody(text, name)
		if !ok {
			continue
		}
		if !sectionQueriesTabsForGo(body) {
			continue
		}
		// Helper must be reachable from boot (or be the body inlined via boot call).
		if !healHelperWiredFromBoot(text, name) && !hasBootRediscoverGoTabs(text) {
			// If boot rediscover is only this helper but unwired, skip.
			continue
		}
		if sectionReconnectsDiscovered(body) {
			return true
		}
		// Softer: body has register/connect after parse (sectionReconnectsDiscovered covers most).
		bl := strings.ToLower(body)
		if strings.Contains(bl, "handleregistermessage") ||
			strings.Contains(bl, "mayberegistergotab") ||
			(strings.Contains(bl, "connectsession") &&
				(strings.Contains(bl, "getorcreatesessionentry") ||
					strings.Contains(bl, "parsegosessionfromurl") ||
					strings.Contains(bl, "sessionid"))) {
			// Reject keys-only connect inside helper.
			if strings.Contains(bl, "sessions.keys") &&
				!strings.Contains(bl, "handleregistermessage") &&
				!strings.Contains(bl, "mayberegistergotab") &&
				!strings.Contains(bl, "getorcreatesessionentry") {
				continue
			}
			return true
		}
	}

	// Inline boot listener: query + reconnect discovered (not keys-only).
	for _, marker := range []string{
		"chrome.runtime.onInstalled",
		"runtime.onInstalled",
		"chrome.runtime.onStartup",
		"runtime.onStartup",
	} {
		sec := listenerNeighborhood(text, marker, 1800)
		if sectionReconnectsDiscovered(sec) {
			return true
		}
		sl := strings.ToLower(sec)
		// Boot calls heal helper that reconnects.
		if hasNamedHealCall(sl) {
			for _, name := range selfHealHelperNames() {
				if !strings.Contains(sl, strings.ToLower(name)) {
					continue
				}
				if body, ok := extractJSFunctionBody(text, name); ok {
					if sectionReconnectsDiscovered(body) {
						return true
					}
					bl := strings.ToLower(body)
					if sectionQueriesTabsForGo(body) &&
						(strings.Contains(bl, "handleregistermessage") ||
							strings.Contains(bl, "mayberegistergotab") ||
							(strings.Contains(bl, "connectsession") &&
								(strings.Contains(bl, "getorcreatesessionentry") ||
									strings.Contains(bl, "parsegosessionfromurl")))) {
						return true
					}
				}
			}
		}
	}
	return false
}

// hasHealReconnectsWithoutDebuggerAttach — heal rediscovers + reconnects WS only.
// Debugger attach inside heal is a popup-UX regression.
func hasHealReconnectsWithoutDebuggerAttach(text string) bool {
	for _, name := range selfHealHelperNames() {
		body, ok := extractJSFunctionBody(text, name)
		if !ok {
			continue
		}
		bl := strings.ToLower(body)
		// Must query tabs / parse go and reconnect.
		if !sectionQueriesTabsForGo(body) {
			continue
		}
		if !(strings.Contains(bl, "mayberegistergotab") ||
			strings.Contains(bl, "connectsession") ||
			strings.Contains(bl, "handleregistermessage")) {
			continue
		}
		// Must NOT attach in heal body (ignore comments: check call-like tokens).
		if strings.Contains(bl, "maybeeagerattach(") ||
			strings.Contains(bl, "attachdebuggerforsession(") ||
			strings.Contains(bl, "listcapturabletabsinsessionwindow(") {
			continue
		}
		return true
	}
	return false
}

// hasReAttachOnHeal — legacy name; contract is now WS-only heal (popup UX).
func hasReAttachOnHeal(text string) bool {
	return hasHealReconnectsWithoutDebuggerAttach(text)
}

// legacy stub: attach-in-heal signals (kept so old references compile if any).
func hasReAttachOnHealLegacyAttachRequired(text string) bool {
	if !hasBootRediscoverGoTabs(text) && !healSectionsHaveRediscover(text) {
		return false
	}

	attachSignals := []string{
		"maybeeagerattach",
		"attachdebuggerforsession",
		"eagerattachtab",
		"eagerattachforsession",
		"autoattachtab",
		"autoattachforsession",
		"maybeattachtabforsession",
		"attachifcapturable",
		"attachcapturabletab",
		"listcapturabletabsinsessionwindow",
	}

	bodyHasAttach := func(bl string) bool {
		for _, a := range attachSignals {
			if strings.Contains(bl, a) {
				return true
			}
		}
		if strings.Contains(bl, "attachdebugger") &&
			(strings.Contains(bl, "session") || strings.Contains(bl, "tabid")) {
			return true
		}
		return false
	}

	for _, name := range selfHealHelperNames() {
		body, ok := extractJSFunctionBody(text, name)
		if !ok {
			continue
		}
		if !healHelperWiredFromBoot(text, name) && !hasBootRediscoverGoTabs(text) {
			continue
		}
		bl := strings.ToLower(body)
		if !bodyHasAttach(bl) {
			continue
		}
		if sectionQueriesTabsForGo(body) ||
			strings.Contains(bl, "connectsession") ||
			strings.Contains(bl, "handleregistermessage") ||
			strings.Contains(bl, "mayberegistergotab") ||
			strings.Contains(bl, "getorcreatesessionentry") {
			return true
		}
	}

	// Boot listener neighborhood: rediscover + attach, or call heal that attaches.
	// Do NOT accept maybeEagerAttach from adjacent onCreated/onUpdated without rediscover.
	for _, marker := range []string{
		"chrome.runtime.onInstalled",
		"runtime.onInstalled",
		"chrome.runtime.onStartup",
		"runtime.onStartup",
	} {
		sec := listenerNeighborhood(text, marker, 2000)
		sl := strings.ToLower(sec)
		if sectionQueriesTabsForGo(sec) && bodyHasAttach(sl) {
			return true
		}
		if hasNamedHealCall(sl) {
			for _, name := range selfHealHelperNames() {
				if !strings.Contains(sl, strings.ToLower(name)) {
					continue
				}
				if body, ok := extractJSFunctionBody(text, name); ok {
					bl := strings.ToLower(body)
					if bodyHasAttach(bl) &&
						(sectionQueriesTabsForGo(body) ||
							strings.Contains(bl, "connectsession") ||
							strings.Contains(bl, "handleregistermessage") ||
							strings.Contains(bl, "mayberegistergotab")) {
						return true
					}
				}
			}
		}
	}
	return false
}

func healSectionsHaveRediscover(text string) bool {
	for _, name := range selfHealHelperNames() {
		if body, ok := extractJSFunctionBody(text, name); ok && sectionQueriesTabsForGo(body) {
			if healHelperWiredFromBoot(text, name) {
				return true
			}
		}
	}
	for _, marker := range []string{
		"chrome.runtime.onInstalled",
		"runtime.onInstalled",
		"chrome.runtime.onStartup",
		"runtime.onStartup",
	} {
		if sectionQueriesTabsForGo(listenerNeighborhood(text, marker, 1800)) {
			return true
		}
	}
	return false
}

// hasHealStrongerThanEmptySessionsKeysLoop — regression: boot must not be only
// for (sessions.keys()) connectSession. Requires tabs.query rediscover of /go + reconnect.
func hasHealStrongerThanEmptySessionsKeysLoop(text string) bool {
	// Positive: boot rediscover of /go tabs + reconnect discovered sessions.
	if !hasBootRediscoverGoTabs(text) {
		return false
	}
	if !hasReconnectWSOnHeal(text) {
		return false
	}
	// hasReconnectWSOnHeal already rejects keys-only connect; this leaf is the
	// explicit regression name for the obsolete pattern.
	return true
}
```
