# browser-agent session eager arm (window = debugging session)

Classic TDD for **P2 eager auto-attach**: when a session window is armed
(at least one open `/go?session=<id>` control tab bound with `windowId`),
capturable tabs in **that window** are **auto-attached** without waiting for a job.

Builds on P1 multi-tab attach set (`attachedTabIds`, policy B) from
`browser-agent-session-attach-gate`. Does **not** re-assert set semantics,
leave detach-all, or attach gate — those stay MECE in the attach-gate tree.

| Surface | What is under test |
|---------|-------------------|
| Extension source | Auto-attach on `/go` register; after `create_tab`; same-window navigate/update; skip non-capturable; other-window not targeted |

**No real Chrome.** L2 static reads of `Chrome-Ext-Browser-Agent/public/background.js`
(also accept `build/` / `src/` fallbacks). Optional L3 e2e is **out of scope** for
this tree (requirement: L2 primary).

## Mode

**Classic TDD.** Current product attaches mainly on **jobs** (`withDebuggerForSession`).
Register only `connectSession`; `createTabInSession` only creates the tab.
Leaves that require eager auto-attach call sites are **RED** until implementer lands P2.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Armed Window** — a Chrome window bound to a session via an open control tab
(`/go?session=<sessionId>`) with `sessions[sessionId].windowId` set.

**Eager Auto-Attach** — call `attachDebuggerForSession(sessionId, tabId)` (or
equivalent session-scoped attach) **proactively** when:

1. Control tab registers (`handleRegisterMessage` / `maybeRegisterGoTab`) — attach
   that control `tabId` when capturable (or always attempt; attach path may no-op
   reuse). Preferred: attach the registering control tab so the window is debugging
   as soon as `/go` is open.
2. `create_tab` / `createTabInSession` creates a new capturable tab in the session
   window — attach the new `tab.id` after `chrome.tabs.create`.
3. Same-window tab open/navigate (`tabs.onUpdated` and/or `tabs.onCreated`) while
   session armed — auto-attach capturable tabs whose `windowId === entry.windowId`.

**Non-capturable skip** — URLs rejected by `isCapturableTabURL` (`chrome://`,
`chrome-extension://`, `devtools://`, `edge://`, `about:`) must **not** be
eager-attached.

**Other-window isolation** — no proactive auto-attach for tabs in windows other
than `entry.windowId`. `createTabInSession` remains scoped to `entry.windowId`.

**Not this tree (MECE / later phases):**

- Multi-attach keep peers, leave detach-all, attach gate, multi-tab recount → P1
  `browser-agent-session-attach-gate`
- Self-heal on SW boot rediscovery → P3
- Popup UX → P4

```text
Open /go?session=S in window W
  -> register binds windowId=W
  -> eager attach control tab (set may include control tabId)

create_tab https://example.com in W
  -> chrome.tabs.create({ windowId: W, ... })
  -> eager attach new tabId (capturable)

User opens/navigates https://app.example in W while armed
  -> tabs.onUpdated (or onCreated)
  -> eager attach if capturable && tab.windowId === W

chrome://settings in W
  -> isCapturableTabURL false -> skip eager attach

Tab in other window W2
  -> not targeted by create or eager attach for session S
```

## Decision Tree

```
browser-agent-session-eager-arm
└── ext-source/                                    [static contract on background.js]
    ├── auto-attach-on-go-register/                  /go register attaches control tab
    ├── auto-attach-on-create-tab/                   create_tab attaches new capturable tab
    ├── auto-attach-on-same-window-navigate/         same-window update/create attaches capturable
    ├── skip-non-capturable/                         chrome:// etc. not auto-attached
    └── other-window-not-targeted/                   create + eager attach scoped to entry.windowId
```

### Parameter significance (high → low)

1. **Trigger** — register vs create_tab vs navigate/update.
2. **Guard** — capturable URL filter vs window scope.
3. **Shared primitive** — all paths should use `attachDebuggerForSession` (policy B set).

## Test Index

| Leaf | Scenario |
|------|----------|
| `ext-source/auto-attach-on-go-register` | `handleRegisterMessage` / `maybeRegisterGoTab` auto-attaches control `tabId` |
| `ext-source/auto-attach-on-create-tab` | After `chrome.tabs.create` in session window, attach new capturable tab |
| `ext-source/auto-attach-on-same-window-navigate` | Armed same-window capturable tab open/navigate → auto-attach |
| `ext-source/skip-non-capturable` | Eager paths guard with `isCapturableTabURL` (skip `chrome://`, etc.) |
| `ext-source/other-window-not-targeted` | Create + eager attach scoped to `entry.windowId`; no other-window attach |

**Leaf count: 5**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-eager-arm
doctest test ./tests/browser-agent-session-eager-arm
```

Static leaves are expected **RED** under job-only attach (current `background.js`
register / create_tab without post-create attach; `tabs.onUpdated` only registers
`/go` and handles leave). Implementer lands eager call sites until GREEN.

```go
import (
	"github.com/xhd2015/doctest/session"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mode — top-level surface under test.
const (
	ModeExtSource = "ext-source"
)

// ExtSourceTarget for ModeExtSource.
const (
	ExtSrcAutoAttachOnGoRegister         = "auto-attach-on-go-register"
	ExtSrcAutoAttachOnCreateTab          = "auto-attach-on-create-tab"
	ExtSrcAutoAttachOnSameWindowNavigate = "auto-attach-on-same-window-navigate"
	ExtSrcSkipNonCapturable              = "skip-non-capturable"
	ExtSrcOtherWindowNotTargeted         = "other-window-not-targeted"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string

	ExtSourceTarget string
}

// Response holds ext-source probe outcomes.
type Response struct {
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	ErrText      string
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

	candidates := shellBackgroundCandidates(root)
	path, data, ok := firstExistingFile(candidates)
	resp.FileExists = ok
	if ok {
		resp.FoundPaths = []string{path}
		resp.FileContents[path] = string(data)
		resp.CombinedText = string(data)
	} else {
		resp.ErrText = "shell background.js not found under Chrome-Ext-Browser-Agent"
	}
	return resp, nil
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
