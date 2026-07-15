# browser-agent session create-tab + Target.* CDP polyfill

Classic-TDD tree for the **create tab** product slice on package
`github.com/xhd2015/browser-agent/browseragent`:

| Surface | What is under test |
|---------|-------------------|
| Go job types | `IsKnownJobType("create_tab")` true (**additive**); unknowns still false; prior six still true |
| React protocol | `react/src/lib/protocol/jobs.ts` exports `create_tab` / `JOB_TYPE_CREATE_TAB` |
| CLI dispatch | `--help` lists `create-tab`; `session create-tab` without session → dual sources |
| CLI job types | HandleCLI posts job type `create_tab` (+ optional `url`); fake WS records first job |
| Extension source | `create_tab` job branch + **all `Target.*` polyfill intercept** via chrome.tabs |
| System prompt | FormatSystemPrompt: create-tab recipe; Target polyfill + **tab_id** (not only Forbidden) |
| SKILL.md | Operator skill documents create-tab + polyfilled Target.* |

**No real Chrome.** Fake WS for job-type observation only. **No real agent-run.**

## Mode

**Classic TDD** — feature not implemented yet. Leaves are expected **RED** until
implementer lands production code. Do **not** implement product source from this tree.

## Sealed / conflict note

Sibling sealed tree `tests/browser-agent-cdp-jobs/` documents **six** job types
(`info|eval|run|logs|screenshot|cdp`). This tree uses **additive** asserts only:

- `create_tab` is known **and** the prior six remain known
- does **not** assert exact set size `len==7` / “exactly seven”

Implementer must expand `IsKnownJobType` / protocol lists additively so
`browser-agent-cdp-jobs` stays GREEN without editing sealed ASSERT files.

If a sealed leaf later hard-fails on set size, orchestrator must approve a
coordinated sealed-tree update — **not** silent edits from implementer.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** / **Agent** creates tabs in the **session window** via CLI or
polyfilled CDP, then jobs against capturable tabs by **tab_id** only.

```text
browser-agent session create-tab [url]
  -> resolve session (--session-id | BROWSER_AGENT_SESSION_ID)
  -> POST /v1/jobs type=create_tab params={url?, active?}
  -> Extension chrome.tabs.create({ windowId: entry.windowId, url?, active? })
  -> result { type: create_tab, tab_id, url?, window_id? }   # NO targetId
```

**Job type** (canonical string, Go + TS):

```text
create_tab
```

| field | type | meaning |
|-------|------|---------|
| `url` | string, optional | omit → blank/new tab |
| `active` | bool, optional | default **true** |

**Public identity:** **`tab_id` only** — responses never require/return CDP
`targetId`. Incoming CDP may still send `targetId` as decimal string of a
chrome tab id (resolved to tab_id).

**CDP polyfill** (extension `handleCdpJob` / pre-hook):

```text
if method starts with "Target."
  -> polyfill dispatch table via chrome.tabs (+ session window rules)
  -> NEVER chrome.debugger.sendCommand for Target.*
else
  -> existing sendDebuggerCommand
```

**Tier A (full polyfill):**

| Method | Behavior |
|--------|----------|
| `Target.createTarget` | shared create path with create_tab job → `{ tab_id, ... }` |
| `Target.closeTarget` | resolve tab → `chrome.tabs.remove`; **reject** closing session-page tab |
| `Target.activateTarget` | `chrome.tabs.update(tabId, { active: true })` |
| `Target.getTargets` | `chrome.tabs.query({ windowId })` → list with **tab_id** |
| `Target.getTargetInfo` | `chrome.tabs.get` → single tab with **tab_id** |

**Tier B (soft):** setDiscoverTargets / setAutoAttach no-op success;
attachToTarget / detachFromTarget map to debugger attach/detach when tab is
in session window.

**Tier C:** other `Target.*` → product error mentioning polyfill unsupported
(not Chrome -32000 "Not allowed").

**Session window rules:**

- Always create/query/close within `entry.windowId` when bound
- `windowId == null` → error: session page not bound / open `/go?session=…`
- Never navigate/close the session control page (`/go?session=<id>`)

**Session resolve** (same as prior trees):

1. `--session-id` flag when set  
2. else env `BROWSER_AGENT_SESSION_ID`  
3. else error mentioning **both** `--session-id` and `BROWSER_AGENT_SESSION_ID`

**HandleCLI** (package API; preferred over binary shell-out):

```text
HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
```

**Test Client** in this tree:

- Dispatch leaves call `HandleCLI` only (empty injectable env).
- Job-type leaves start `browseragent.Run` (NoOpenChrome, NoAgentRun) + **fake WS**
  that records the **first** job envelope (`type` + `params`) then auto-completes.
- System / go-job-types pure package calls.
- Ext-source / protocol-src / skill-md read **ModuleRoot** filesystem.

## Decision Tree

```
browser-agent-create-tab-target-polyfill
├── go-job-types/                              [IsKnownJobType — additive]
│   └── known-create-tab/                        G1 create_tab true; prior six true; unknown false
├── protocol-src/                              [react jobs.ts]
│   └── jobs-module-create-tab/                  P1 create_tab / JOB_TYPE_CREATE_TAB present
├── cli-dispatch/                              [HandleCLI only]
│   ├── help-lists-create-tab/                   A1 --help lists create-tab under session +\n
│   └── create-tab-without-session/              A2 session create-tab missing session → flag + env
├── cli-job-types/                             [serve + fake WS + nested HandleCLI]
│   ├── create-tab-blank/                        B1 type=create_tab; optional empty url
│   └── create-tab-with-url/                     B2 type=create_tab; params include url
├── ext-source/                                [filesystem tokens — shell background]
│   ├── shell-create-tab-branch/                 E1 create_tab job branch + chrome.tabs.create
│   ├── shell-target-intercept/                  E2 Target. prefix intercept; not raw sendCommand only
│   ├── shell-target-tier-a/                     E3 Tier A method tokens + chrome.tabs APIs
│   ├── shell-target-session-rules/              E4 windowId scope; session-page protect; tab_id identity
│   └── shell-target-soft-unsupported/           E5 Tier B soft + Tier C unsupported polyfill error
├── system-prompt/                             [FormatSystemPrompt]
│   ├── create-tab-recipe/                       C1 nested session create-tab recipe
│   └── target-polyfill/                         C2 Target polyfilled + tab_id; not Forbidden-only
└── skill-md/                                  [SKILL.md filesystem]
    └── create-tab-target/                       S1 create-tab + Target polyfill language
```

### Parameter significance (high → low)

1. **Surface / Mode** — go helper vs protocol FS vs CLI dispatch vs live job vs
   extension FS vs system prompt vs skill doc (different `Run` branches).
2. **Within CLI** — help (discoverability) vs missing-session (error path) vs
   live job success (type + params).
3. **Job params** — blank create-tab vs create-tab with URL.
4. **Ext-source concern** — job branch vs intercept gate vs Tier A methods vs
   session rules / identity vs soft+unsupported.
5. **Docs concern** — create-tab CLI recipe vs Target polyfill narrative.

## Test Index

| Leaf | Scenario |
|------|----------|
| `go-job-types/known-create-tab` | (G1) `IsKnownJobType("create_tab")` true; prior six true; unknowns false (**additive**, no exact count) |
| `protocol-src/jobs-module-create-tab` | (P1) jobs.ts contains `create_tab` / `JOB_TYPE_CREATE_TAB` (quoted preferred) |
| `cli-dispatch/help-lists-create-tab` | (A1) `--help` nil err; lists `create-tab` (or `session create-tab`); trailing `\n` |
| `cli-dispatch/create-tab-without-session` | (A2) `session create-tab` no session → err mentions `--session-id` + `BROWSER_AGENT_SESSION_ID` |
| `cli-job-types/create-tab-blank` | (B1) serve+fake WS; `session create-tab --session-id …` → observed type `create_tab`; CLI ok + `\n` |
| `cli-job-types/create-tab-with-url` | (B2) `session create-tab <url>` → type `create_tab`; params include url |
| `ext-source/shell-create-tab-branch` | (E1) shell background: `create_tab` job branch + `chrome.tabs.create` |
| `ext-source/shell-target-intercept` | (E2) Target.* intercept path; not only fall-through to raw sendCommand |
| `ext-source/shell-target-tier-a` | (E3) Target.createTarget/closeTarget/activateTarget/getTargets/getTargetInfo + chrome.tabs.* |
| `ext-source/shell-target-session-rules` | (E4) windowId / session-window scope; protect `/go?session`; results use `tab_id` |
| `ext-source/shell-target-soft-unsupported` | (E5) soft Target methods and/or polyfill unsupported error language (light) |
| `system-prompt/create-tab-recipe` | (C1) FormatSystemPrompt includes `browser-agent session create-tab` |
| `system-prompt/target-polyfill` | (C2) Target polyfill + tab_id language; not Forbidden Target.* only |
| `skill-md/create-tab-target` | (S1) SKILL.md documents create-tab and polyfilled Target.* / tab_id |

**Leaf count: 14**

## How to Run

```sh
cd browser-agent   # module root
doctest vet ./tests/browser-agent-create-tab-target-polyfill
doctest test ./tests/browser-agent-create-tab-target-polyfill
# expect RED until implementer lands feature

# regressions (sealed — must stay GREEN after additive implement)
doctest test ./tests/browser-agent-cdp-jobs/...
doctest test ./tests/browser-agent-session-nested/...
doctest test ./tests/browser-agent-session-tab-targeting/...
```

Module: `github.com/xhd2015/browser-agent`.  
Package under test: `…/browseragent`.

### Implementer contract (authoritative for GREEN)

**CLI**

```go
func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
```

- Nested: `session create-tab [url]` (positional URL enough; optional `--url` OK if also accepted).
- Posts job type **`create_tab`** with params `{ "url"?: string, "active"?: bool }` (active default true).
- Missing session → error text includes `--session-id` and `BROWSER_AGENT_SESSION_ID`.
- Help lists `create-tab` under session subcommands.
- Successful stdout ends with `\n`.

**Go job type helper (additive)**

```go
const JobTypeCreateTab = "create_tab"
func IsKnownJobType(s string) bool  // must accept create_tab AND prior six
```

Do **not** remove prior six. Prefer not asserting exact cardinality in any tree.

**React protocol**

```text
react/src/lib/protocol/jobs.ts
# JOB_TYPE_CREATE_TAB = "create_tab"
# KNOWN_JOB_TYPES includes create_tab
```

**Extension** (`Chrome-Ext-Browser-Agent/**/background.js` preferred; embed mirror if required)

- `handleJob` case `"create_tab"` (shared create path with Target.createTarget).
- CDP: if method starts with `Target.` → polyfill table; never raw `chrome.debugger.sendCommand` for Target.*.
- Tier A methods implemented via chrome.tabs; identity **`tab_id`**.
- Session window rules + never close/navigate session page.
- Tier B soft no-op / attach map; Tier C unsupported polyfill error (not -32000 Not allowed).

**Docs**

- `FormatSystemPrompt`: recipe `browser-agent session create-tab`; Target.* polyfilled (tab lifecycle) + tab_id; replace blanket Forbidden-only wording.
- `browseragent/SKILL.md` (and/or `cmd/browser-agent/SKILL.md`): create-tab + polyfill language.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode values — top-level surface under test.
const (
	ModeCLIDispatch  = "cli-dispatch"
	ModeCLIJobTypes  = "cli-job-types"
	ModeSystemPrompt = "system-prompt"
	ModeExtSource    = "ext-source"
	ModeProtocolSrc  = "protocol-src"
	ModeGoJobTypes   = "go-job-types"
	ModeSkillMD      = "skill-md"
)

// DispatchKind for ModeCLIDispatch.
const (
	DispatchHelp                    = "help"
	DispatchCreateTabWithoutSession = "create-tab-without-session"
)

// JobCLI for ModeCLIJobTypes.
const (
	JobCLICreateTab = "create-tab"
)

// ExtSourceTarget for ModeExtSource.
const (
	ExtSrcShellCreateTabBranch     = "shell-create-tab-branch"
	ExtSrcShellTargetIntercept     = "shell-target-intercept"
	ExtSrcShellTargetTierA         = "shell-target-tier-a"
	ExtSrcShellTargetSessionRules  = "shell-target-session-rules"
	ExtSrcShellTargetSoftUnsup     = "shell-target-soft-unsupported"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// ModuleRoot is module directory (filesystem leaves).
	// Root Setup sets from DOCTEST_ROOT/../..
	ModuleRoot string

	// BaseDir is temp parent for serve fixtures.
	BaseDir string

	// Session / server (cli-job-types)
	Addr         string
	SessionID    string
	NoOpenChrome bool
	NoAgentRun   bool
	ReadyTimeout time.Duration

	// --- CLI shared ---
	CLIArgs         []string
	CLIEnv          map[string]string
	MaxDispatchWait time.Duration

	// DispatchKind for ModeCLIDispatch.
	DispatchKind string

	// --- cli-job-types ---
	JobCLI string
	// CreateTabURL optional positional/flag URL for create-tab.
	CreateTabURL string
	// FakeExtension enables recording WS (default true for job-types).
	FakeExtension bool

	// --- system-prompt ---
	PromptSessionID string

	// --- ext-source ---
	ExtSourceTarget string

	// --- skill-md ---
	SkillMDProbe string
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// CLI shared
	Stdout           string
	Stderr           string
	ExitCode         int
	ErrText          string
	CLIErr           string
	DispatchTimedOut bool

	// Observed job from fake WS (first job only)
	ObservedJobType   string
	ObservedJobParams map[string]any
	ObservedJobRaw    string
	WSJobReceived     bool
	JobsSeen          int

	// Server meta
	BaseURL       string
	RealSessionID string

	// System prompt
	SystemPrompt string

	// Filesystem probes
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string

	// Go job types
	KnownResults    map[string]bool
	UnknownResults  map[string]bool
	HelperAvailable bool

	// SKILL.md
	SkillFileExists bool
	SkillPath       string
	SkillText       string
	SkillPathsTried []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}
	switch req.Mode {
	case ModeCLIDispatch:
		return runCLIDispatch(t, req)
	case ModeCLIJobTypes:
		return runCLIJobTypes(t, req)
	case ModeSystemPrompt:
		return runSystemPrompt(t, req)
	case ModeExtSource:
		return runExtSource(t, req)
	case ModeProtocolSrc:
		return runProtocolSrc(t, req)
	case ModeGoJobTypes:
		return runGoJobTypes(t, req)
	case ModeSkillMD:
		return runSkillMD(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- CLI dispatch (no server) ---

func runCLIDispatch(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.DispatchKind == "" && len(req.CLIArgs) == 0 {
		t.Fatal("DispatchKind or CLIArgs must be set")
	}
	if len(req.CLIArgs) == 0 {
		switch req.DispatchKind {
		case DispatchHelp:
			req.CLIArgs = []string{"--help"}
		case DispatchCreateTabWithoutSession:
			req.CLIArgs = []string{"session", "create-tab"}
		default:
			return nil, fmt.Errorf("unknown DispatchKind %q", req.DispatchKind)
		}
	}
	return invokeHandleCLI(t, req)
}

func invokeHandleCLI(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	maxWait := req.MaxDispatchWait
	if maxWait <= 0 {
		maxWait = 3 * time.Second
	}
	var stdout, stderr bytes.Buffer
	env := req.CLIEnv
	if env == nil {
		env = map[string]string{}
	}
	args := req.CLIArgs
	if args == nil {
		args = []string{}
	}

	type outcome struct {
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		done <- outcome{err: browseragent.HandleCLI(args, env, &stdout, &stderr)}
	}()

	resp := &Response{}
	select {
	case out := <-done:
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		if out.err != nil {
			resp.CLIErr = out.err.Error()
			resp.ErrText = out.err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		resp.ExitCode = 0
		return resp, nil
	case <-time.After(maxWait):
		resp.DispatchTimedOut = true
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.ExitCode = 1
		resp.ErrText = "HandleCLI timed out (possible accidental serve hang)"
		return resp, fmt.Errorf("%s", resp.ErrText)
	}
}

// --- CLI job types against live server + recording fake WS ---

func runCLIJobTypes(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.JobCLI == "" {
		t.Fatal("JobCLI must be set (create-tab)")
	}
	srv, cleanup, err := startAgentServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		BaseURL:       srv.BaseURL,
		RealSessionID: srv.SessionID,
	}

	ext, err := dialFakeExtension(srv.BaseURL, "1.0.0", []string{"browser-agent"})
	if err != nil {
		return resp, fmt.Errorf("fake extension dial: %w", err)
	}
	defer ext.Close()

	var mu sync.Mutex
	ext.OnJob = func(env wsEnvelope) {
		mu.Lock()
		defer mu.Unlock()
		if resp.WSJobReceived {
			return
		}
		resp.WSJobReceived = true
		resp.ObservedJobType = envelopeJobType(env)
		resp.ObservedJobParams = envelopeJobParams(env)
		b, _ := json.Marshal(env)
		resp.ObservedJobRaw = string(b)
	}
	ext.AutoCompleteOK = true
	ext.ResultData = map[string]any{
		"type":      "create_tab",
		"tab_id":    float64(12345),
		"url":       "about:blank",
		"window_id": float64(1),
	}
	go ext.Loop()
	time.Sleep(40 * time.Millisecond)

	args := req.CLIArgs
	if len(args) == 0 {
		args = buildJobCLIArgs(req, srv)
	} else {
		args = injectAddrAndSession(args, srv.BaseURL, srv.SessionID)
	}

	req2 := *req
	req2.CLIArgs = args
	if req2.CLIEnv == nil {
		req2.CLIEnv = map[string]string{}
	}
	if req2.MaxDispatchWait <= 0 {
		req2.MaxDispatchWait = 10 * time.Second
	}
	cliResp, err := invokeHandleCLI(t, &req2)
	if cliResp != nil {
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
		resp.CLIErr = cliResp.CLIErr
		resp.ErrText = cliResp.ErrText
		resp.DispatchTimedOut = cliResp.DispatchTimedOut
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		got := resp.WSJobReceived
		mu.Unlock()
		if got {
			break
		}
		time.Sleep(15 * time.Millisecond)
	}
	mu.Lock()
	resp.JobsSeen = ext.JobsSeen
	mu.Unlock()
	return resp, err
}

func buildJobCLIArgs(req *Request, srv *agentServer) []string {
	sid := srv.SessionID
	addr := srv.BaseURL
	switch req.JobCLI {
	case JobCLICreateTab:
		args := []string{"session", "create-tab", "--session-id", sid, "--addr", addr}
		if strings.TrimSpace(req.CreateTabURL) != "" {
			args = append(args, req.CreateTabURL)
		}
		return args
	default:
		return []string{"session", req.JobCLI, "--session-id", sid, "--addr", addr}
	}
}

func injectAddrAndSession(args []string, baseURL, sessionID string) []string {
	hasAddr, hasSession := false, false
	for _, a := range args {
		if a == "--addr" || strings.HasPrefix(a, "--addr=") {
			hasAddr = true
		}
		if a == "--session-id" || strings.HasPrefix(a, "--session-id=") {
			hasSession = true
		}
	}
	out := append([]string{}, args...)
	insertAt := 0
	if len(out) >= 2 && out[0] == "session" {
		insertAt = 2
	} else if len(out) >= 1 {
		insertAt = 1
	}
	if !hasSession {
		ins := []string{"--session-id", sessionID}
		out = append(out[:insertAt:insertAt], append(ins, out[insertAt:]...)...)
		insertAt += 2
	}
	if !hasAddr {
		ins := []string{"--addr", baseURL}
		out = append(out[:insertAt:insertAt], append(ins, out[insertAt:]...)...)
	}
	return out
}

// --- system prompt ---

func runSystemPrompt(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	sid := req.PromptSessionID
	if sid == "" {
		sid = "sess-create-tab-prompt"
	}
	return &Response{
		SystemPrompt:  browseragent.FormatSystemPrompt(sid),
		RealSessionID: sid,
	}, nil
}

// --- extension source filesystem ---

func runExtSource(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ExtSourceTarget == "" {
		t.Fatal("ExtSourceTarget must be set")
	}
	root := req.ModuleRoot
	resp := &Response{FileContents: map[string]string{}}

	switch req.ExtSourceTarget {
	case ExtSrcShellCreateTabBranch,
		ExtSrcShellTargetIntercept,
		ExtSrcShellTargetTierA,
		ExtSrcShellTargetSessionRules,
		ExtSrcShellTargetSoftUnsup:
		candidates := []string{
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "background.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "background.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "background.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "background.js"),
		}
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
	default:
		return nil, fmt.Errorf("unknown ExtSourceTarget %q", req.ExtSourceTarget)
	}
}

// --- react protocol jobs module ---

func runProtocolSrc(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	root := req.ModuleRoot
	reactRoot := filepath.Join(root, "react")
	if st, err := os.Stat(reactRoot); err != nil || !st.IsDir() {
		alt := filepath.Join(root, "project-api-capture-react")
		if st2, err2 := os.Stat(alt); err2 == nil && st2.IsDir() {
			reactRoot = alt
		}
	}
	candidates := []string{
		filepath.Join(reactRoot, "src", "lib", "protocol", "jobs.ts"),
		filepath.Join(reactRoot, "src", "lib", "protocol", "jobs.tsx"),
		filepath.Join(reactRoot, "src", "lib", "protocol", "jobs.js"),
		filepath.Join(reactRoot, "src", "protocol", "jobs.ts"),
		filepath.Join(reactRoot, "src", "protocol", "jobs.js"),
	}
	path, data, ok := firstExistingFile(candidates)
	resp := &Response{FileContents: map[string]string{}}
	resp.FileExists = ok
	if ok {
		resp.FoundPaths = []string{path}
		resp.FileContents[path] = string(data)
		resp.CombinedText = string(data)
	} else {
		resp.ErrText = "react protocol jobs module not found under react/src/lib/protocol/jobs.*"
	}
	return resp, nil
}

// --- Go IsKnownJobType (additive create_tab) ---

func runGoJobTypes(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	// Additive: prior six + create_tab. Do NOT assert exclusive set size.
	known := []string{"info", "eval", "run", "logs", "screenshot", "cdp", "create_tab"}
	unknown := []string{"", "foo", "Eval", "navigate", "job", "unknown-type", "create-tab", "CreateTab", "target"}
	resp := &Response{
		KnownResults:    map[string]bool{},
		UnknownResults:  map[string]bool{},
		HelperAvailable: true,
	}
	for _, k := range known {
		resp.KnownResults[k] = browseragent.IsKnownJobType(k)
	}
	for _, u := range unknown {
		resp.UnknownResults[u] = browseragent.IsKnownJobType(u)
	}
	return resp, nil
}

// --- SKILL.md ---

func runSkillMD(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SkillMDProbe == "" {
		// default probe name for single-leaf skill group
		req.SkillMDProbe = "create-tab-target"
	}
	root := req.ModuleRoot
	candidates := []string{
		filepath.Join(root, "browseragent", "SKILL.md"),
		filepath.Join(root, "cmd", "browser-agent", "SKILL.md"),
	}
	resp := &Response{SkillPathsTried: candidates}
	path, data, ok := firstExistingFile(candidates)
	resp.SkillFileExists = ok
	if ok {
		resp.SkillPath = path
		resp.SkillText = string(data)
		resp.CombinedText = string(data)
		resp.FoundPaths = []string{path}
		resp.FileExists = true
	} else {
		resp.ErrText = "SKILL.md not found under browseragent/ or cmd/browser-agent/"
	}
	return resp, nil
}

// --- server harness ---

type agentServer struct {
	BaseURL   string
	SessionID string
	cancel    context.CancelFunc
	done      <-chan error
}

func startAgentServer(t *testing.T, req *Request) (*agentServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	if req.SessionID == "" {
		t.Fatal("SessionID must be set by Setup")
	}
	addr := req.Addr
	if addr == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, nil, err
		}
		addr = ln.Addr().String()
		_ = ln.Close()
		req.Addr = addr
	}
	ready := req.ReadyTimeout
	if ready <= 0 {
		ready = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	var stdout, stderr bytes.Buffer
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      req.BaseDir,
		SessionID:    req.SessionID,
		NoOpenChrome: true,
		NoAgentRun:   true,
		Stdout:       &stdout,
		Stderr:       &stderr,
	}

	done := make(chan error, 1)
	go func() {
		_, err := browseragent.Run(ctx, cfg)
		done <- err
	}()

	baseURL := "http://" + addr
	if err := waitHealth(baseURL, ready); err != nil {
		cancel()
		<-done
		return nil, nil, fmt.Errorf("control server never healthy at %s: %w", baseURL, err)
	}

	srv := &agentServer{
		BaseURL:   baseURL,
		SessionID: req.SessionID,
		cancel:    cancel,
		done:      done,
	}
	cleanup := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
	}
	return srv, cleanup, nil
}

func waitHealth(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/health", nil)
		if err != nil {
			cancel()
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err == nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			cancel()
			if res.StatusCode == http.StatusOK {
				return nil
			}
			last = fmt.Errorf("health status %d", res.StatusCode)
		} else {
			last = err
			cancel()
		}
		time.Sleep(20 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("timeout waiting for health")
	}
	return last
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

// --- fake extension WS client (records first job) ---

type wsEnvelope struct {
	V       int            `json:"v"`
	Type    string         `json:"type"`
	ID      string         `json:"id,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

type fakeExtension struct {
	conn           *websocket.Conn
	AutoCompleteOK bool
	ResultData     map[string]any
	OnJob          func(wsEnvelope)
	JobsSeen       int
	version        string
	features       []string
	mu             sync.Mutex
}

func dialFakeExtension(baseURL, version string, features []string) (*fakeExtension, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
	dialer := websocket.Dialer{HandshakeTimeout: 2 * time.Second}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	if version == "" {
		version = "1.0.0"
	}
	if features == nil {
		features = []string{"browser-agent"}
	}
	return &fakeExtension{
		conn:     conn,
		version:  version,
		features: features,
	}, nil
}

func (f *fakeExtension) Close() {
	if f != nil && f.conn != nil {
		_ = f.conn.Close()
	}
}

func (f *fakeExtension) SendHello() error {
	env := wsEnvelope{
		V:    1,
		Type: "hello",
		Payload: map[string]any{
			"version":  f.version,
			"features": f.features,
		},
	}
	return f.conn.WriteJSON(env)
}

func (f *fakeExtension) Loop() {
	_ = f.SendHello()
	for {
		var env wsEnvelope
		if err := f.conn.ReadJSON(&env); err != nil {
			return
		}
		if env.Type == "job" {
			f.mu.Lock()
			f.JobsSeen++
			f.mu.Unlock()
			if f.OnJob != nil {
				f.OnJob(env)
			}
			if f.AutoCompleteOK {
				jobID := env.ID
				if jobID == "" && env.Payload != nil {
					if id, ok := env.Payload["id"].(string); ok {
						jobID = id
					} else if id, ok := env.Payload["job_id"].(string); ok {
						jobID = id
					}
				}
				data := f.ResultData
				if data == nil {
					data = map[string]any{"tab_id": 1}
				}
				_ = f.conn.WriteJSON(wsEnvelope{
					V:    1,
					Type: "result",
					ID:   jobID,
					Payload: map[string]any{
						"job_id": jobID,
						"ok":     true,
						"data":   data,
					},
				})
			}
		}
	}
}

func envelopeJobType(env wsEnvelope) string {
	if env.Payload == nil {
		return ""
	}
	if t, ok := env.Payload["type"].(string); ok && t != "" {
		return t
	}
	if t, ok := env.Payload["job_type"].(string); ok && t != "" {
		return t
	}
	if job, ok := env.Payload["job"].(map[string]any); ok {
		if t, ok := job["type"].(string); ok {
			return t
		}
	}
	return ""
}

func envelopeJobParams(env wsEnvelope) map[string]any {
	if env.Payload == nil {
		return nil
	}
	if p, ok := env.Payload["params"].(map[string]any); ok {
		return p
	}
	if job, ok := env.Payload["job"].(map[string]any); ok {
		if p, ok := job["params"].(map[string]any); ok {
			return p
		}
	}
	out := map[string]any{}
	for k, v := range env.Payload {
		if k == "type" || k == "id" || k == "job_id" || k == "job_type" {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
```
