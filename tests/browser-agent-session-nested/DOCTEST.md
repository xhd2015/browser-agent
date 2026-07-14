# browser-agent nested `session` CLI + agent-run env/prefix + SYSTEM.md without control id

Exercises the **complete refactor** of browser-agent side-commands under
`session`, agent-run id prefix + `--env`, and SYSTEM.md playbook without a
concrete control session id.

Package: `github.com/xhd2015/browser-agent/browseragent`

| Surface | What is under test |
|---------|-------------------|
| AgentRunSessionID | `browser-agent-sess-` + control; idempotent when already prefixed |
| BuildAgentRunArgs | `--session-id=<agent-run-id>`, `--env BROWSER_AGENT_SESSION_ID=<control>`, `--no-submit` |
| FormatSystemPrompt | nested recipes `browser-agent session …`; **no** concrete control id |
| HandleCLI nested | top-level `session`; flat side cmds fail; help lists `session` |

**No production code in this tree.** Leaves stay **RED** until implementer lands
nested CLI + args. Related sealed trees (`browser-agent*`) are updated in place
for the same contract.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** runs **`browser-agent`** with two command families:

1. **Top-level** — `serve`, `install-chrome-extension`, `skill` (and help).
2. **Nested session side-commands** — only under **`session`**:
   `info | eval | run | logs | screenshot | cdp`.

```text
browser-agent session info|eval|run|logs|screenshot|cdp …
browser-agent serve | install-chrome-extension | skill
```

Flat `browser-agent info|eval|…` is **not** a side-command handler (unknown /
brief error). There are **no** flat aliases after the complete refactor.

**Control id** vs **agent-run id**:

```text
control id   = serve session id (disk, jobs, logs, BROWSER_AGENT_SESSION_ID value)
agent-run id = AgentRunSessionID(control)
             = "browser-agent-sess-" + control
             (if control already has prefix AgentRunSessionIDPrefix, return as-is)

AgentRunSessionIDPrefix = "browser-agent-sess-"
```

**BuildAgentRunArgs(controlSessionID, promptOrSystemPath, workspaceDir) []string**

Canonical tokens (order flexible except prompt after `--`):

```text
run
--session-id=<AgentRunSessionID(control)>     # agent-run id, not bare control
--agent-runner=grok-tty
--auto-send-or-resume
--new-terminal
[--dir=<workspace>]                           # only when workspace non-empty
--env BROWSER_AGENT_SESSION_ID=<control>      # control id for CLI resolve
--no-submit                                   # ALWAYS draft open
--open
--
<open prompt>                                 # must not require bare control id
```

Production **launchAgentRun** must use this argv; **no** manual `cmd.Env`
overlay for session id (session env is the `--env` flag). Serve operator logs
still print the **control** session id.

**Session resolve** (side-commands under `session`):

1. `--session-id` flag when set  
2. else env `BROWSER_AGENT_SESSION_ID`  
3. else error mentioning **both** `--session-id` and `BROWSER_AGENT_SESSION_ID`

**SYSTEM.md / FormatSystemPrompt**:

- Recipes use nested form: `browser-agent session info|eval|run|logs|screenshot|cdp`
- Body must **not** embed a concrete control session id argument
- Mentions `BROWSER_AGENT_SESSION_ID` as env source for resolve

**Test Client** in this tree:

- Pure leaves call `AgentRunSessionID`, `BuildAgentRunArgs`, `FormatSystemPrompt`
- CLI leaves call `HandleCLI` with injectable env + buffers
- `session-eval-posts-job` starts in-process serve + fake WS, then nested eval

## Decision Tree

```
browser-agent-session-nested
├── agent-run-id/                              [pure AgentRunSessionID]
│   ├── bare-control/                            A1 demo → browser-agent-sess-demo
│   └── already-prefixed/                        A2 no double prefix
├── agent-run-args/                            [pure BuildAgentRunArgs]
│   ├── core-flags/                              A3 prefix session-id + --env + --no-submit
│   └── open-prompt-no-control-id/               A4 open prompt omits bare control id
├── system-prompt/                             [FormatSystemPrompt]
│   ├── nested-recipes/                          B1 browser-agent session info|eval|…
│   ├── no-control-id/                           B2 unique control id absent from body
│   └── env-source-mention/                      B3 BROWSER_AGENT_SESSION_ID mentioned
└── cli-nested/                                [HandleCLI]
    ├── help-lists-session/                      C1 --help lists session + nested cmds
    ├── session-info-without-session/            C2 session info → flag+env error
    ├── flat-info-unknown/                       C3 flat info → unknown / not success
    └── session-eval-posts-job/                  C4 session eval posts type=eval (fake WS)
```

### Parameter significance (high → low)

1. **Surface / Mode** — id pure vs argv pure vs system prompt vs CLI (different
   `Run` contracts).
2. **Within pure id** — bare control vs already-prefixed (idempotency).
3. **Within argv** — core flags vs open-prompt content.
4. **Within system prompt** — recipes vs absence of id vs env mention.
5. **Within CLI** — help vs missing session vs flat unknown vs live eval.

## Test Index

| Leaf | Scenario |
|------|----------|
| `agent-run-id/bare-control` | (A1) control `demo` → `browser-agent-sess-demo` |
| `agent-run-id/already-prefixed` | (A2) already `browser-agent-sess-…` → no double prefix |
| `agent-run-args/core-flags` | (A3) argv: `--session-id=browser-agent-sess-…`, `--env BROWSER_AGENT_SESSION_ID=<control>`, `--no-submit` |
| `agent-run-args/open-prompt-no-control-id` | (A4) SYSTEM.md open prompt does not contain bare control id |
| `system-prompt/nested-recipes` | (B1) body has `browser-agent session info` (+ eval/run/logs/screenshot/cdp) |
| `system-prompt/no-control-id` | (B2) unique control id **not** present in body |
| `system-prompt/env-source-mention` | (B3) mentions `BROWSER_AGENT_SESSION_ID` |
| `cli-nested/help-lists-session` | (C1) `--help` lists `session` and nested cmds; nil err; `\n` |
| `cli-nested/session-info-without-session` | (C2) `session info` no sid → err mentions flag + env |
| `cli-nested/flat-info-unknown` | (C3) flat `info` → error / unknown (not successful handler) |
| `cli-nested/session-eval-posts-job` | (C4) `session eval` with session posts job type `eval` (fake WS) |

**Leaf count: 11**

## How to Run

```sh
cd project-api-capture
doctest vet ./tests/browser-agent-session-nested
doctest test ./tests/browser-agent-session-nested/...
# after implement, also:
# doctest test ./tests/browser-agent...
# doctest test ./tests/browser-agent-cli-react/...
# doctest test ./tests/browser-agent-cdp-jobs/...
# doctest test ./tests/browser-agent-serve-runtime/...
```

### Implementer contract (authoritative for GREEN)

```text
const AgentRunSessionIDPrefix = "browser-agent-sess-"

func AgentRunSessionID(controlID string) string
// if control already has prefix → return as-is
// else return prefix + control

func BuildAgentRunArgs(controlSessionID, promptOrSystemPath, workspaceDir string) []string
// --session-id=<AgentRunSessionID(control)>
// --env BROWSER_AGENT_SESSION_ID=<control>   (flag form: --env K=V or --env K V)
// --no-submit always
// open prompt must not require embedding bare control id

func FormatSystemPrompt(sessionID string) string
// may accept sessionID for API stability but MUST NOT embed concrete control id
// recipes: browser-agent session info|eval|run|logs|screenshot|cdp
// mention BROWSER_AGENT_SESSION_ID

func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
// top-level: serve | session | install-chrome-extension | skill | help
// session sub: info|eval|run|logs|screenshot|cdp
// flat info/eval/… → unknown / non-success (no side-command handler)
```

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
	ModeAgentRunID    = "agent-run-id"
	ModeAgentRunArgs  = "agent-run-args"
	ModeSystemPrompt  = "system-prompt"
	ModeCLINested     = "cli-nested"
)

// CLIKind for ModeCLINested.
const (
	CLIKindHelp                    = "help"
	CLIKindSessionInfoNoSession    = "session-info-without-session"
	CLIKindFlatInfoUnknown         = "flat-info-unknown"
	CLIKindSessionEvalPostsJob     = "session-eval-posts-job"
)

// AgentRunSessionIDPrefix is the locked product prefix (tests assert this literal).
const AgentRunSessionIDPrefix = "browser-agent-sess-"

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// --- agent-run-id ---
	ControlSessionID string

	// --- agent-run-args ---
	AgentArgsControlID  string
	AgentArgsPromptPath string
	AgentArgsWorkspace  string

	// --- system-prompt ---
	PromptSessionID string

	// --- cli-nested ---
	CLIKind         string
	CLIArgs         []string
	CLIEnv          map[string]string
	MaxDispatchWait time.Duration

	// Live serve (session-eval)
	BaseDir      string
	Addr         string
	SessionID    string
	NoOpenChrome bool
	NoAgentRun   bool
	ReadyTimeout time.Duration
	EvalExpr     string
	FakeExtension bool
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// Pure id / args
	AgentRunID   string
	AgentRunArgs []string
	OpenPrompt   string // prompt token(s) after "--" when present
	ControlID    string

	// System prompt
	SystemPrompt string

	// CLI
	Stdout           string
	Stderr           string
	ExitCode         int
	ErrText          string
	CLIErr           string
	DispatchTimedOut bool

	// Live eval observation
	BaseURL           string
	RealSessionID     string
	ObservedJobType   string
	ObservedJobParams map[string]any
	WSJobReceived     bool
	JobsSeen          []map[string]any
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeAgentRunID:
		return runAgentRunID(t, req)
	case ModeAgentRunArgs:
		return runAgentRunArgs(t, req)
	case ModeSystemPrompt:
		return runSystemPrompt(t, req)
	case ModeCLINested:
		return runCLINested(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- pure AgentRunSessionID ---

func runAgentRunID(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	ctrl := req.ControlSessionID
	if ctrl == "" {
		t.Fatal("ControlSessionID must be set")
	}
	id := browseragent.AgentRunSessionID(ctrl)
	return &Response{
		ControlID:  ctrl,
		AgentRunID: id,
		ExitCode:   0,
	}, nil
}

// --- pure BuildAgentRunArgs ---

func runAgentRunArgs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	ctrl := req.AgentArgsControlID
	if ctrl == "" {
		ctrl = "ctrl-agent-args"
	}
	prompt := req.AgentArgsPromptPath
	if prompt == "" {
		prompt = "/tmp/sessions/" + ctrl + "/SYSTEM.md"
	}
	args := browseragent.BuildAgentRunArgs(ctrl, prompt, req.AgentArgsWorkspace)
	return &Response{
		ControlID:    ctrl,
		AgentRunArgs: args,
		OpenPrompt:   extractOpenPrompt(args),
		AgentRunID:   browseragent.AgentRunSessionID(ctrl),
		ExitCode:     0,
	}, nil
}

func extractOpenPrompt(args []string) string {
	for i, a := range args {
		if a == "--" && i+1 < len(args) {
			return strings.Join(args[i+1:], " ")
		}
	}
	// fallback: last non-flag-looking token after open-ish region
	if len(args) == 0 {
		return ""
	}
	return args[len(args)-1]
}

// --- FormatSystemPrompt ---

func runSystemPrompt(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	sid := req.PromptSessionID
	if sid == "" {
		sid = "sess-nested-prompt"
	}
	return &Response{
		ControlID:    sid,
		SystemPrompt: browseragent.FormatSystemPrompt(sid),
		ExitCode:     0,
	}, nil
}

// --- CLI nested ---

func runCLINested(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLIKind == "" && len(req.CLIArgs) == 0 {
		t.Fatal("CLIKind or CLIArgs must be set")
	}
	if len(req.CLIArgs) == 0 {
		switch req.CLIKind {
		case CLIKindHelp:
			req.CLIArgs = []string{"--help"}
		case CLIKindSessionInfoNoSession:
			req.CLIArgs = []string{"session", "info"}
		case CLIKindFlatInfoUnknown:
			req.CLIArgs = []string{"info"}
		case CLIKindSessionEvalPostsJob:
			// filled after server start
		default:
			return nil, fmt.Errorf("unknown CLIKind %q", req.CLIKind)
		}
	}
	if req.CLIKind == CLIKindSessionEvalPostsJob {
		return runSessionEvalPostsJob(t, req)
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

func runSessionEvalPostsJob(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	req.NoOpenChrome = true
	req.NoAgentRun = true
	req.FakeExtension = true
	if req.MaxDispatchWait <= 0 {
		req.MaxDispatchWait = 10 * time.Second
	}
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}

	srv, cleanup, err := startAgentServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var mu sync.Mutex
	resp := &Response{
		BaseURL:       srv.BaseURL,
		RealSessionID: srv.SessionID,
		ControlID:     srv.SessionID,
	}

	ext, err := dialFakeExtension(srv.BaseURL, "1.0.0", []string{"browser-agent"})
	if err != nil {
		return resp, fmt.Errorf("fake extension dial: %w", err)
	}
	defer ext.Close()
	ext.AutoCompleteOK = true
	ext.ResultData = map[string]any{"value": 2, "result": 2, "ok": true}
	ext.OnJob = func(jobType string, params map[string]any) {
		mu.Lock()
		defer mu.Unlock()
		resp.WSJobReceived = true
		resp.ObservedJobType = jobType
		resp.ObservedJobParams = params
		resp.JobsSeen = append(resp.JobsSeen, map[string]any{
			"type":   jobType,
			"params": params,
		})
	}
	go ext.Loop()
	time.Sleep(40 * time.Millisecond)

	expr := req.EvalExpr
	if expr == "" {
		expr = "1+1"
	}
	args := req.CLIArgs
	if len(args) == 0 {
		args = []string{
			"session", "eval",
			"--session-id", srv.SessionID,
			"--addr", srv.BaseURL,
			expr,
		}
	} else {
		args = injectAddrAndSession(args, srv.BaseURL, srv.SessionID)
	}

	req2 := *req
	req2.CLIArgs = args
	cliResp, err := invokeHandleCLI(t, &req2)
	if cliResp != nil {
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
		resp.CLIErr = cliResp.CLIErr
		resp.ErrText = cliResp.ErrText
		resp.DispatchTimedOut = cliResp.DispatchTimedOut
	}
	// Settle observer
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
	return resp, err
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

// --- serve + fake extension (minimal, shared with other browser-agent trees) ---

type agentServer struct {
	BaseURL   string
	SessionID string
	BaseDir   string
	cancel    context.CancelFunc
}

func startAgentServer(t *testing.T, req *Request) (*agentServer, func(), error) {
	t.Helper()
	baseDir := req.BaseDir
	if baseDir == "" {
		var err error
		baseDir, err = os.MkdirTemp("", "ba-session-nested-*")
		if err != nil {
			return nil, nil, err
		}
	}
	sid := req.SessionID
	if sid == "" {
		sid = fmt.Sprintf("sess-nested-%d", time.Now().UnixNano()%1e12)
	}
	addr := req.Addr
	if addr == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, nil, err
		}
		addr = ln.Addr().String()
		_ = ln.Close()
	}
	readyTO := req.ReadyTimeout
	if readyTO <= 0 {
		readyTO = 5 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      baseDir,
		SessionID:    sid,
		NoOpenChrome: true,
		NoAgentRun:   true,
		ReadyTimeout: readyTO,
		Stdout:       io.Discard,
		Stderr:       io.Discard,
	}
	errCh := make(chan error, 1)
	go func() {
		_, err := browseragent.Run(ctx, cfg)
		errCh <- err
	}()
	baseURL := "http://" + addr
	if err := waitHealth(baseURL, readyTO); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("serve health: %w", err)
	}
	srv := &agentServer{BaseURL: baseURL, SessionID: sid, BaseDir: baseDir, cancel: cancel}
	cleanup := func() {
		cancel()
		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
		}
		_ = os.RemoveAll(baseDir)
	}
	return srv, cleanup, nil
}

func waitHealth(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		resp, err := http.Get(strings.TrimRight(baseURL, "/") + "/v1/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
			last = fmt.Errorf("health status %d", resp.StatusCode)
		} else {
			last = err
		}
		time.Sleep(25 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("health timeout")
	}
	return last
}

type fakeExtension struct {
	conn            *websocket.Conn
	AutoCompleteOK  bool
	ResultData      map[string]any
	OnJob           func(jobType string, params map[string]any)
	JobsSeen        []map[string]any
	mu              sync.Mutex
}

func dialFakeExtension(baseURL, version string, features []string) (*fakeExtension, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	ext := &fakeExtension{conn: conn}
	hello := map[string]any{
		"v":    1,
		"type": "hello",
		"payload": map[string]any{
			"version":  version,
			"features": features,
		},
	}
	if err := conn.WriteJSON(hello); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return ext, nil
}

func (f *fakeExtension) Close() {
	if f.conn != nil {
		_ = f.conn.Close()
	}
}

func (f *fakeExtension) Loop() {
	for {
		var msg map[string]any
		if err := f.conn.ReadJSON(&msg); err != nil {
			return
		}
		typ, _ := msg["type"].(string)
		if typ != "job" {
			continue
		}
		payload, _ := msg["payload"].(map[string]any)
		if payload == nil {
			payload, _ = msg["job"].(map[string]any)
		}
		jobID := stringField(payload, "id", "job_id", "jobId")
		if jobID == "" {
			jobID = stringField(msg, "id", "job_id", "jobId")
		}
		jobType := stringField(payload, "type", "job_type", "jobType")
		params, _ := payload["params"].(map[string]any)
		f.mu.Lock()
		f.JobsSeen = append(f.JobsSeen, map[string]any{"type": jobType, "params": params})
		onJob := f.OnJob
		f.mu.Unlock()
		if onJob != nil {
			onJob(jobType, params)
		}
		if f.AutoCompleteOK {
			result := map[string]any{
				"v":    1,
				"type": "result",
				"payload": map[string]any{
					"job_id": jobID,
					"ok":     true,
					"data":   f.ResultData,
				},
			}
			_ = f.conn.WriteJSON(result)
		}
	}
}

func stringField(m map[string]any, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			switch t := v.(type) {
			case string:
				if t != "" {
					return t
				}
			default:
				s := fmt.Sprint(t)
				if s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}

// silence unused imports in some codegen paths
var (
	_ = bytes.Buffer{}
	_ = json.Marshal
	_ = filepath.Join
	_ = sync.Mutex{}
)
```
