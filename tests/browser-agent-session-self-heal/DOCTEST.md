# browser-agent session self-heal (SW boot rediscover)

Classic TDD for **P3 self-heal on SW boot**: after `onInstalled` / cold MV3 service
worker restart, in-memory `sessions` is empty. The boot path must **rediscover**
open `/go?session=` control tabs via `chrome.tabs.query`, rebind session +
`windowId`, reconnect WS, and re-run eager attach for control + capturable tabs
in that window.

Builds on P1 multi-tab attach (`browser-agent-session-attach-gate`) and P2 eager
arm (`browser-agent-session-eager-arm`). Does **not** re-assert multi-attach set
semantics, leave detach-all, or per-trigger eager arm call sites — those stay
MECE in their trees. This tree only asserts heal **invokes** rediscover + reconnect
+ re-attach on boot.

| Surface | What is under test |
|---------|-------------------|
| Extension source | Boot rediscover of `/go` tabs; reconnect WS; re-attach; stronger than empty `sessions.keys()` loop |

**No real Chrome.** L2 static reads of `Chrome-Ext-Browser-Agent/public/background.js`
(also accept `build/` / `src/` fallbacks). E2e / popup (P4) out of scope.

## Mode

**Classic TDD.** Current product `onInstalled` / `onStartup` only loop
`sessions.keys()` then `connectSession` — empty after cold SW boot, so open
`/go` tabs are never rebound. Leaves that require tabs.query rediscover + re-arm
are **RED** until implementer lands P3.

## Version

0.0.1

# DSN (Domain Specific Notion)

**Cold SW Boot** — MV3 service worker restarts (install, browser restart, idle
eviction). Module-level `sessions` Map and `sessionAttachState` start empty.

**Self-Heal / Boot Rediscover** — on `chrome.runtime.onInstalled`,
`chrome.runtime.onStartup`, and/or top-level SW init, query open tabs for
`/go?session=` (via `chrome.tabs.query` + `parseGoSessionFromURL` /
`isSessionGoPageURL` / equivalent), then for each discovered control tab:

1. **Register/bind** — `getOrCreateSessionEntry` / `handleRegisterMessage` /
   `maybeRegisterGoTab` / equivalent so `sessions[sessionId]` has `tabId` +
   `windowId`.
2. **Reconnect WS** — `connectSession(sessionId, …)` for the discovered id
   (not only ids already in the empty map).
3. **Re-arm attach** — `maybeEagerAttach` / `attachDebuggerForSession` (or
   equivalent) for the control tab and/or capturable tabs in that window
   (reuses P2 helpers; do not change multi-attach or eager semantics beyond
   invoking them on heal).

**Empty-keys insufficiency** — `for (const sessionId of sessions.keys())
connectSession(sessionId, …)` alone is the **obsolete** boot path: after cold
restart `sessions` is empty, so nothing reconnects or re-attaches. Heal must be
stronger (tab query rediscover).

**Optional** — `chrome.storage` for session ids is fine but not required if tab
query alone rediscovers open `/go` pages.

**Not this tree (MECE / later phases):**

- Multi-attach keep peers, leave detach-all, attach gate → P1
  `browser-agent-session-attach-gate`
- Eager triggers on register / create_tab / same-window navigate → P2
  `browser-agent-session-eager-arm`
- Popup UX → P4

```text
SW cold start; open tab http://127.0.0.1:43761/go?session=S in window W
  -> onInstalled / onStartup / init
  -> chrome.tabs.query({}) (or equivalent)
  -> parse /go?session=S from tab URL
  -> bind sessions[S] = { tabId, windowId: W, ... }
  -> connectSession(S, "heal"|"onInstalled"|…)
  -> maybeEagerAttach / attachDebuggerForSession for control + capturable in W

Obsolete (insufficient):
  for (sessionId of sessions.keys()) connectSession(...)  // empty map → no-op
```

## Decision Tree

```
browser-agent-session-self-heal
└── ext-source/                                    [static contract on background.js]
    ├── boot-rediscover-go-tabs/                     boot queries tabs + parses /go?session=
    ├── reconnect-ws-on-heal/                        heal connects WS for discovered sessions
    ├── re-attach-on-heal/                           heal re-arms attach for window tabs
    └── empty-sessions-keys-insufficient/            regression: empty keys loop alone fails
```

### Parameter significance (high → low)

1. **Rediscover** — tabs.query + parse `/go?session=` at boot (foundation).
2. **Reconnect** — bind + `connectSession` for discovered ids.
3. **Re-attach** — invoke eager/session attach after rebind.
4. **Regression** — prove heal is stronger than obsolete empty-keys loop.

## Test Index

| Leaf | Scenario |
|------|----------|
| `ext-source/boot-rediscover-go-tabs` | Boot path (`onInstalled` / `onStartup` / heal helper) calls `chrome.tabs.query` and parses `/go?session=` |
| `ext-source/reconnect-ws-on-heal` | Heal path registers/binds discovered sessions and calls `connectSession` (not empty-map loop only) |
| `ext-source/re-attach-on-heal` | Heal path calls `maybeEagerAttach` / `attachDebuggerForSession` for control and/or capturable tabs |
| `ext-source/empty-sessions-keys-insufficient` | Regression: boot heal must not be only `for (sessions.keys()) connectSession` |

**Leaf count: 4**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-self-heal
doctest test ./tests/browser-agent-session-self-heal
```

Static leaves are expected **RED** under the empty `sessions.keys()` boot loop
(current `background.js`). Implementer lands tabs.query rediscover + rebind +
reconnect + re-attach until GREEN.

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
	ExtSrcBootRediscoverGoTabs            = "boot-rediscover-go-tabs"
	ExtSrcReconnectWSOnHeal               = "reconnect-ws-on-heal"
	ExtSrcReAttachOnHeal                  = "re-attach-on-heal"
	ExtSrcEmptySessionsKeysInsufficient   = "empty-sessions-keys-insufficient"
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
