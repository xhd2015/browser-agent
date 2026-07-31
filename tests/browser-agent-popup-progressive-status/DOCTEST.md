# browser-agent popup progressive status (instant shell + phases)

Classic TDD for **P4 instant popup + progressive status**: toolbar popup opens
with an **immediate shell** (version + layout), then surfaces connect/attach
progress through clear phases (daemon → session/WS → debugger). Prefer
**last-known** from storage for first status paint; live health/status runs
**non-blocking** with timeout. Never block first paint on network/SW.

Builds on P1 multi-tab attach, P2 eager arm, P3 self-heal (background.js). Does
**not** re-assert attach gate, eager triggers, or boot rediscover — those stay
MECE in their trees. This tree only asserts **popup UX progressive status** and
the SW **status API / last-known storage** surfaces the popup needs.

| Surface | What is under test |
|---------|-------------------|
| `popup.html` | Instant version shell; progressive phase DOM hooks |
| `popup.js` | Sync version paint; health timeout; last-known storage read |
| `background.js` | Status message handler and/or storage last-known write |

**No real Chrome.** L2 static reads of `Chrome-Ext-Browser-Agent/public/`
(also accept `build/` / `src/` fallbacks). L3 e2e / badge optional — out of scope.

## Mode

**Classic TDD.** Current popup is a small HTML shell + async `fetch` health only:
no AbortController/timeout, no session/WS/debugger phase rows, no last-known
storage, no SW status API for the popup. Leaves that require progressive phases,
timeout, SW status, or last-known paint are **RED** until implementer lands P4.

Shell version identity (sync `bundle-sum` → `#pkg-version`) may already **GREEN**.

## Version

0.0.1

# DSN (Domain Specific Notion)

**Toolbar Popup** — MV3 `action.default_popup` (`popup.html` + `popup.js`). Opens
on toolbar click. Must paint **shell** (title, package version/md5) immediately
via HTML and/or synchronous `bundle-sum.js` — never wait on `/v1/health` or SW
round-trip for first paint of identity.

**Progressive Status** — status UI with distinct **phases** (not a single hung
`checking…` with no timeout):

1. **Daemon / control** — HTTP `/v1/health` (or control port reachability)
2. **Session / WS** — extension session WebSocket connected or waiting
3. **Debugger / armed** — attach set / multi-attach friendly armed state

Each phase has a DOM hook (id / data attribute / dedicated row) so the UI can
update independently as background work completes.

**Health with timeout** — popup health/status `fetch` uses `AbortController` (or
equivalent timeout/abort) so failures surface as unreachable/error, not infinite
checking.

**Last-known** — SW writes session armed + debugger info to
`chrome.storage.session` (or `local`); popup **reads** last-known before or in
parallel with live refresh for first status paint.

**SW status API** — `chrome.runtime.onMessage` type `status` / `getStatus` /
`popupStatus` (or equivalent) and/or storage keys the popup can query, including
session armed + debugger/attach info.

**Not this tree (MECE / earlier phases):**

- Multi-attach keep peers, leave detach-all, attach gate → P1
  `browser-agent-session-attach-gate`
- Eager triggers on register / create_tab / navigate → P2
  `browser-agent-session-eager-arm`
- Boot rediscover self-heal → P3 `browser-agent-session-self-heal`
- Side panel, full visual redesign, optional badge

```text
User opens toolbar popup
  -> shell paints version/md5 (bundle-sum sync / HTML ids) immediately
  -> last-known from chrome.storage.session|local paints status phases (optional stale)
  -> parallel: fetch /v1/health with AbortController timeout
  -> parallel: SW getStatus / storage refresh for session+WS + debugger armed
  -> UI updates phase rows: daemon | session/ws | debugger
  -> timeout/error -> phase shows unreachable/error (not infinite checking…)
```

## Decision Tree

```
browser-agent-popup-progressive-status
└── ext-source/                                    [static contract on popup + background]
    ├── shell-version-instant/                       version shell without network wait
    ├── health-fetch-timeout/                        AbortController/timeout on health fetch
    ├── progressive-status-phases/                   DOM hooks for daemon, session/ws, debugger
    ├── sw-status-api/                               background status message and/or storage write
    └── last-known-status-paint/                     popup reads storage last-known status
```

### Parameter significance (high → low)

1. **Contract surface** — shell identity vs health timeout vs phase DOM vs SW API
   vs last-known paint (five independent P4 exit criteria).
2. **File surface** — popup.html / popup.js / background.js (asserted per leaf).
3. **Status dimension** — daemon vs session/ws vs debugger (phases leaf only).

## Test Index

| Leaf | Scenario |
|------|----------|
| `ext-source/shell-version-instant` | `#pkg-version` (or equivalent) + sync version paint; not only after fetch |
| `ext-source/health-fetch-timeout` | Health/status fetch uses AbortController or timeout abort |
| `ext-source/progressive-status-phases` | Distinct DOM hooks/rows for daemon, session/ws, debugger phases |
| `ext-source/sw-status-api` | Background onMessage status query and/or storage last-known write with session+debugger |
| `ext-source/last-known-status-paint` | Popup reads chrome.storage session/local last status before or parallel with live |

**Leaf count: 5**

## How to Run

```sh
doctest vet ./tests/browser-agent-popup-progressive-status
doctest test ./tests/browser-agent-popup-progressive-status
```

Expected under **current** popup (health-only, no timeout/phases/status API):

| Leaf | Expected |
|------|----------|
| `shell-version-instant` | **GREEN** (already has `#pkg-version` + sync `BROWSER_AGENT_BUNDLE_VERSION`) |
| `health-fetch-timeout` | **RED** (bare `fetch` without abort/timeout) |
| `progressive-status-phases` | **RED** (only `#ctrl-status` / single checking…) |
| `sw-status-api` | **RED** (onMessage handles `register` only; no popup status storage) |
| `last-known-status-paint` | **RED** (popup does not read `chrome.storage`) |

Implementer lands progressive UI + timeout + SW status/last-known until remaining leaves GREEN.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

// Mode — top-level surface under test.
const (
	ModeExtSource = "ext-source"
)

// ExtSourceTarget for ModeExtSource.
const (
	ExtSrcShellVersionInstant      = "shell-version-instant"
	ExtSrcHealthFetchTimeout       = "health-fetch-timeout"
	ExtSrcProgressiveStatusPhases  = "progressive-status-phases"
	ExtSrcSWStatusAPI              = "sw-status-api"
	ExtSrcLastKnownStatusPaint     = "last-known-status-paint"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string

	ExtSourceTarget string
}

// Response holds ext-source probe outcomes for popup + background files.
type Response struct {
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	// Named convenience fields (may be empty when file missing).
	PopupHTML string
	PopupJS   string
	BackgroundJS string
	ErrText   string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}

	switch req.Mode {
	case ModeExtSource:
		return runExtSource(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runExtSource(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ExtSourceTarget == "" {
		t.Fatal("ExtSourceTarget must be set")
	}
	root := req.ModuleRoot
	resp := &Response{FileContents: map[string]string{}}

	// Always probe popup.html, popup.js, background.js so leaves can pick what they need.
	htmlPath, htmlData, htmlOK := firstExistingFile(shellPopupHTMLCandidates(root))
	jsPath, jsData, jsOK := firstExistingFile(shellPopupJSCandidates(root))
	bgPath, bgData, bgOK := firstExistingFile(shellBackgroundCandidates(root))

	var parts []string
	if htmlOK {
		resp.FoundPaths = append(resp.FoundPaths, htmlPath)
		resp.FileContents[htmlPath] = string(htmlData)
		resp.PopupHTML = string(htmlData)
		parts = append(parts, string(htmlData))
	}
	if jsOK {
		resp.FoundPaths = append(resp.FoundPaths, jsPath)
		resp.FileContents[jsPath] = string(jsData)
		resp.PopupJS = string(jsData)
		parts = append(parts, string(jsData))
	}
	if bgOK {
		resp.FoundPaths = append(resp.FoundPaths, bgPath)
		resp.FileContents[bgPath] = string(bgData)
		resp.BackgroundJS = string(bgData)
		parts = append(parts, string(bgData))
	}

	resp.CombinedText = strings.Join(parts, "\n")
	resp.FileExists = htmlOK || jsOK || bgOK

	// Target-specific required files.
	switch req.ExtSourceTarget {
	case ExtSrcShellVersionInstant, ExtSrcProgressiveStatusPhases:
		if !htmlOK && !jsOK {
			resp.ErrText = "popup.html / popup.js not found under Chrome-Ext-Browser-Agent"
			resp.FileExists = false
		}
	case ExtSrcHealthFetchTimeout, ExtSrcLastKnownStatusPaint:
		if !jsOK {
			resp.ErrText = "popup.js not found under Chrome-Ext-Browser-Agent"
			resp.FileExists = false
		}
	case ExtSrcSWStatusAPI:
		if !bgOK {
			resp.ErrText = "background.js not found under Chrome-Ext-Browser-Agent"
			resp.FileExists = false
		}
	}

	return resp, nil
}

func shellPopupHTMLCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "popup.html"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "popup.html"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "popup.html"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "popup.html"),
	}
}

func shellPopupJSCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "popup.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "popup.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "popup.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "popup.js"),
	}
}

func shellBackgroundCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "background.js"),
	}
}

func firstExistingFile(paths []string) (string, []byte, bool) {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return p, data, true
		}
	}
	return "", nil, false
}

// Silence unused import if helpers live only in SETUP.md.
var _ = strings.Contains
```
