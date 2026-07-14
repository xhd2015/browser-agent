# browser-agent CDP job types + full CLI side-commands

Exercises the **CDP-oriented job surface** on package
`github.com/xhd2015/browser-agent/browseragent` (next slice after
sealed MVP + CLI/React + serve-runtime trees):

| Surface | What is under test |
|---------|-------------------|
| CLI dispatch | `--help` lists `session` + nested cmds; `session run|logs|screenshot|cdp` missing session |
| CLI job types | HandleCLI nested posts correct `POST /v1/jobs` type + params; fake WS records first job |
| SYSTEM.md recipes | `FormatSystemPrompt` nested `browser-agent session …` recipes (no control id) |
| Extension CDP source | shell + embedded background: `chrome.debugger` / CDP method tokens + job branches |
| React protocol module | `react/src/lib/protocol/jobs.ts` job type string constants |
| Go known job types | `IsKnownJobType` (or equivalent) accepts six types; rejects unknown |

**No real Chrome.** Fake WS for job-type observation only. **No real agent-run.**

**Sealed** (must stay GREEN; **do not modify**):

```sh
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-cli-react/...
doctest test ./tests/browser-agent-serve-runtime/...
```

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** / **Agent** runs **`browser-agent`** CLI **nested** side-commands
under `session` that all resolve a **session id** then talk to the **Control
Server**:

```text
browser-agent serve
browser-agent session info
browser-agent session eval '<expr>'
browser-agent session run <file.js>
browser-agent session logs
browser-agent session screenshot [-o path]
browser-agent session cdp <Method> [json-params]
```

Flat `info|eval|run|logs|screenshot|cdp` are **not** side-command handlers.

**Session resolve** (same rules as prior trees):

1. `--session-id` flag when set  
2. else env `BROWSER_AGENT_SESSION_ID`  
3. else error mentioning **both** `--session-id` and `BROWSER_AGENT_SESSION_ID`

**HandleCLI** (package API; preferred over binary shell-out):

```text
HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
```

- `--help` / `-h` → full help listing **serve, session** and nested
  **info, eval, run, logs, screenshot, cdp**; nil error; trailing `\n`
- nested side-commands without session → non-nil error naming both sources
- successful side-commands print result JSON-ish + trailing `\n`
- `--addr` points at control server (tests use free port base URL)

**Job types** (canonical strings, shared Go + optional TS):

```text
info | eval | run | logs | screenshot | cdp
```

| type | CLI (nested) | params (HTTP body) |
|------|--------------|--------------------|
| eval | `session eval <expr>` | `{ "expression": "..." }` (expr alias OK) |
| run | `session run <path.js>` | `{ "source": "<file contents>" }` (expression OK alt) |
| logs | `session logs` | optional `{ "limit", "level" }` |
| screenshot | `session screenshot [-o]` | optional `{ "format":"png", "full_page":false }` |
| cdp | `session cdp <Method> [json]` | `{ "method":"Page.navigate", "params":{...} }` |
| info | `session info` | `{}` (may also be GET `/v1/session`; extension job still named info) |

**Extension Agent** (shell + embedded mini) handles jobs via **CDP-oriented**
code (not pure stub comments):

- `chrome.debugger` attach / sendCommand / detach pattern (shell preferred)
- `Runtime.evaluate` for eval/run
- `Page.captureScreenshot` for screenshot
- branches for `eval`, `run`, `logs`, `screenshot`, `cdp`, `info`
- mini embed may be thinner but must **mention** CDP method names + job type tokens

**Shared protocol** (preferred):

- Go: `IsKnownJobType(s string) bool` and/or exported type string constants
- TS: `react/src/lib/protocol/jobs.ts` (or `.js`) with the six type strings

**Test Client** in this tree:

- Dispatch leaves call `HandleCLI` only (empty injectable env).
- Job-type leaves start `browseragent.Run` (NoOpenChrome, NoAgentRun) + **fake WS**
  that records the **first** job envelope (`type` + `params`) then auto-completes.
- System / go-job-types pure package calls.
- Ext-source / protocol-src read **ModuleRoot** filesystem.

## Decision Tree

```
browser-agent-cdp-jobs
├── cli-dispatch/                              [HandleCLI only]
│   ├── help/                                    A1 --help lists session + nested cmds +\n
│   ├── run-without-session/                     A2 session run missing session → flag + env
│   ├── logs-without-session/                    A3 session logs missing session
│   ├── screenshot-without-session/              A4 session screenshot missing session
│   └── cdp-without-session/                     A5 session cdp missing session
├── cli-job-types/                             [serve + fake WS + nested HandleCLI]
│   ├── eval/                                    B1 type=eval; expression observed
│   ├── run/                                     B2 type=run; source/file content
│   ├── logs/                                    B3 type=logs
│   ├── screenshot/                              B4 type=screenshot
│   └── cdp/                                     B5 type=cdp; method Page.navigate
├── system-prompt/                             [FormatSystemPrompt]
│   └── format-contains-recipes/                 C1 nested session recipes; no control id
├── ext-source/                                [filesystem CDP tokens]
│   ├── shell-cdp-tokens/                        D1 debugger + Runtime.evaluate + captureScreenshot
│   ├── shell-job-branches/                      D2 eval,run,logs,screenshot,cdp,info
│   └── embedded-cdp-tokens/                     D3 Runtime.evaluate + job type tokens
├── protocol-src/                              [react protocol jobs module]
│   └── jobs-module/                             E1 jobs.ts with six type constants
└── go-job-types/                              [IsKnownJobType / constants]
    └── known-types/                             F1 six known true; unknown false
```

### Parameter significance (high → low)

1. **Surface / Mode** — dispatch vs live job types vs prompt vs FS source vs
   protocol vs Go helper (different `Run` branches).
2. **Within CLI** — help vs missing-session (error path) vs live job type success.
3. **Job type** — eval / run / logs / screenshot / cdp (MECE canonical set for
   this tree; `info` covered by prior cli-react sidecmd + extension branch leaf).
4. **Ext-source target** — shell CDP methods vs shell job branches vs embedded mini.
5. **Leaf details** — exact param keys, method name, file content markers.

## Test Index

| Leaf | Scenario |
|------|----------|
| `cli-dispatch/help` | (A1) `--help` → nil err; lists `serve`, `session`, nested info/eval/run/logs/screenshot/cdp; trailing `\n` |
| `cli-dispatch/run-without-session` | (A2) `session run x.js` no session → err mentions `--session-id` + `BROWSER_AGENT_SESSION_ID` |
| `cli-dispatch/logs-without-session` | (A3) `session logs` no session → same dual mention |
| `cli-dispatch/screenshot-without-session` | (A4) `session screenshot` no session → same |
| `cli-dispatch/cdp-without-session` | (A5) `session cdp Page.navigate` no session → same |
| `cli-job-types/eval` | (B1) serve+fake WS; `session eval --session-id … '1+1'` → observed type `eval` + expression; CLI ok + `\n` |
| `cli-job-types/run` | (B2) temp `.js`; `session run` → type `run`; params include source/file content; CLI ok + `\n` |
| `cli-job-types/logs` | (B3) `session logs` → type `logs`; CLI ok + `\n` |
| `cli-job-types/screenshot` | (B4) `session screenshot` → type `screenshot`; CLI ok + `\n` |
| `cli-job-types/cdp` | (B5) `session cdp Page.navigate {"url":"https://example.com"}` → type `cdp`; method `Page.navigate` |
| `system-prompt/format-contains-recipes` | (C1) FormatSystemPrompt nested session recipes; no concrete control id |
| `ext-source/shell-cdp-tokens` | (D1) shell background: `chrome.debugger` + `Runtime.evaluate` + `Page.captureScreenshot` |
| `ext-source/shell-job-branches` | (D2) shell background mentions job types eval, run, logs, screenshot, cdp, info |
| `ext-source/embedded-cdp-tokens` | (D3) embedded mini: at least `Runtime.evaluate` + job type tokens |
| `protocol-src/jobs-module` | (E1) `react/src/lib/protocol/jobs.{ts,js}` exists with all six type strings |
| `go-job-types/known-types` | (F1) known six → true; unknown → false via `IsKnownJobType` |

**Leaf count: 16**

## How to Run

```sh
cd project-api-capture
doctest vet ./tests/browser-agent-cdp-jobs
doctest test ./tests/browser-agent-cdp-jobs/...
# or:
cd tests/browser-agent-cdp-jobs && doctest vet . && doctest test -v .

# regressions (sealed)
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-cli-react/...
doctest test ./tests/browser-agent-serve-runtime/...
```

Module: `github.com/xhd2015/browser-agent`.  
Package under test: `…/browseragent`.

### Implementer contract (authoritative for GREEN)

**CLI — new / extended commands**

```go
func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
```

- Help lists: `serve`, `session`, and nested `info`, `eval`, `run`, `logs`,
  `screenshot`, `cdp`.
- `session run [--session-id] [--addr] <path.js>` — read file → POST type=`run`
  params with `source` (file body). `expression` alt acceptable if documented.
- `session logs [--session-id] [--addr] [--limit N]` — POST type=`logs`.
- `session screenshot [--session-id] [--addr] [-o file.png]` — POST type=`screenshot`;
  optional write base64 to `-o` (not required for this tree’s asserts).
- `session cdp [--session-id] [--addr] <Method> [json]` — POST type=`cdp` with
  `params.method` + optional nested `params.params`.
- `session eval` continues to POST type=`eval` with expression.
- Missing session → error text includes `--session-id` and `BROWSER_AGENT_SESSION_ID`.
- Successful stdout ends with `\n`.

**Go job type helper**

```go
func IsKnownJobType(s string) bool
// Optional constants:
// JobTypeInfo, JobTypeEval, JobTypeRun, JobTypeLogs, JobTypeScreenshot, JobTypeCDP
```

Known: `info`, `eval`, `run`, `logs`, `screenshot`, `cdp` (exact lowercase).

**Extension sources**

- `Chrome-Ext-Browser-Agent/**/background.js` — CDP-oriented handlers
- `browseragent/embedded/extension/background.js` — thinner OK; CDP method +
  job type name tokens required

**React protocol (preferred path)**

```text
react/src/lib/protocol/jobs.ts   # or .js / .tsx
# must contain string tokens: info, eval, run, logs, screenshot, cdp
```

**FormatSystemPrompt** — nested `browser-agent session` recipes for info, eval,
run, logs, screenshot, cdp; **no** concrete control session id; mentions
`BROWSER_AGENT_SESSION_ID`
(already partially present; ensure all six appear as CLI recipes).

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
)

// DispatchKind for ModeCLIDispatch.
const (
	DispatchHelp                 = "help"
	DispatchRunWithoutSession    = "run-without-session"
	DispatchLogsWithoutSession   = "logs-without-session"
	DispatchShotWithoutSession   = "screenshot-without-session"
	DispatchCDPWithoutSession    = "cdp-without-session"
)

// JobCLI for ModeCLIJobTypes — which side-command to invoke.
const (
	JobCLIEval       = "eval"
	JobCLIRun        = "run"
	JobCLILogs       = "logs"
	JobCLIScreenshot = "screenshot"
	JobCLICdp        = "cdp"
)

// ExtSourceTarget for ModeExtSource.
const (
	ExtSrcShellCDPTokens    = "shell-cdp-tokens"
	ExtSrcShellJobBranches  = "shell-job-branches"
	ExtSrcEmbeddedCDPTokens = "embedded-cdp-tokens"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// ModuleRoot is project-api-capture module directory (filesystem leaves).
	// Root Setup sets from DOCTEST_ROOT/../..
	ModuleRoot string

	// BaseDir is temp parent for serve / run script fixtures.
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
	// EvalExpr for eval (default "1+1").
	EvalExpr string
	// RunScriptPath / RunScriptBody for run command.
	RunScriptPath string
	RunScriptBody string
	// CDPMethod / CDPParamsJSON for cdp command.
	CDPMethod     string
	CDPParamsJSON string
	// FakeExtension enables recording WS (default true for job-types).
	FakeExtension bool

	// --- system-prompt ---
	PromptSessionID string

	// --- ext-source ---
	ExtSourceTarget string
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
	KnownResults   map[string]bool
	UnknownResults map[string]bool
	// HelperAvailable is true when IsKnownJobType is callable (always after implement).
	HelperAvailable bool
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
		case DispatchRunWithoutSession:
			// nested session path; relative script; never reaches file read if session fails first
			req.CLIArgs = []string{"session", "run", "script.js"}
		case DispatchLogsWithoutSession:
			req.CLIArgs = []string{"session", "logs"}
		case DispatchShotWithoutSession:
			req.CLIArgs = []string{"session", "screenshot"}
		case DispatchCDPWithoutSession:
			req.CLIArgs = []string{"session", "cdp", "Page.navigate", `{"url":"https://example.com"}`}
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
		t.Fatal("JobCLI must be set (eval|run|logs|screenshot|cdp)")
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

	// Recording fake extension: capture first job, then auto-complete.
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
		"value":  2,
		"ok":     true,
		"result": map[string]any{"stub": true},
		"base64": "iVBORw0KGgo=",
		"format": "png",
		"entries": []any{},
	}
	go ext.Loop()
	time.Sleep(40 * time.Millisecond)

	// Materialize run script if needed.
	if req.JobCLI == JobCLIRun {
		body := req.RunScriptBody
		if body == "" {
			body = "// doctest-run-marker\nconsole.log('hello-from-run');\n"
		}
		path := req.RunScriptPath
		if path == "" {
			path = filepath.Join(req.BaseDir, "script.js")
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return resp, err
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return resp, err
		}
		req.RunScriptPath = path
		req.RunScriptBody = body
	}

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
	// Allow WS observer to settle after CLI returns.
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
	case JobCLIEval:
		expr := req.EvalExpr
		if expr == "" {
			expr = "1+1"
		}
		return []string{"session", "eval", "--session-id", sid, "--addr", addr, expr}
	case JobCLIRun:
		return []string{"session", "run", "--session-id", sid, "--addr", addr, req.RunScriptPath}
	case JobCLILogs:
		return []string{"session", "logs", "--session-id", sid, "--addr", addr}
	case JobCLIScreenshot:
		return []string{"session", "screenshot", "--session-id", sid, "--addr", addr}
	case JobCLICdp:
		method := req.CDPMethod
		if method == "" {
			method = "Page.navigate"
		}
		params := req.CDPParamsJSON
		if params == "" {
			params = `{"url":"https://example.com"}`
		}
		return []string{"session", "cdp", "--session-id", sid, "--addr", addr, method, params}
	default:
		// Prefer nested form for unknown JobCLI tokens that look like side-commands.
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
		sid = "sess-cdp-prompt"
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
	case ExtSrcShellCDPTokens, ExtSrcShellJobBranches:
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

	case ExtSrcEmbeddedCDPTokens:
		candidates := []string{
			filepath.Join(root, "browseragent", "embedded", "extension", "background.js"),
			filepath.Join(root, "browseragent", "embedded", "extension", "service_worker.js"),
			filepath.Join(root, "browseragent", "embedded", "extension", "sw.js"),
		}
		path, data, ok := firstExistingFile(candidates)
		if !ok && req.BaseDir != "" {
			if installPath, _, err := browseragent.ExtractEmbeddedExtension(req.BaseDir); err == nil {
				extractCandidates := []string{
					filepath.Join(installPath, "background.js"),
					filepath.Join(installPath, "service_worker.js"),
					filepath.Join(installPath, "sw.js"),
				}
				path, data, ok = firstExistingFile(extractCandidates)
			}
		}
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "embedded extension background not found"
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

// --- Go IsKnownJobType ---

func runGoJobTypes(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	known := []string{"info", "eval", "run", "logs", "screenshot", "cdp"}
	unknown := []string{"", "foo", "Eval", "navigate", "job", "unknown-type"}
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
					data = map[string]any{"value": 2}
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
	// Some servers nest job under payload.job
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
	// Fallback: whole payload minus type/id fields
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
