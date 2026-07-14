# browser-agent daemon Phase 8 — session new + EnsureDaemon

Phase 8 adds **`browser-agent session new [--session-id XXX]`** — ensure the
daemon is running, create a session via `POST /v1/sessions`, open Chrome via an
injectable hook, and print operator-facing stdout. **Never** launches agent-run.

Package APIs:

- **`EnsureDaemon(cfg EnsureDaemonConfig) (DaemonMeta, error)`** — health reuse or
  spawn detached serve, wait for healthy + `server.json`
- **`SessionNew(cfg SessionNewConfig) error`** — ensure daemon, create session,
  `OpenChromeFn`, pretty stdout
- **`HandleCLI session new`** — CLI dispatch via `cliSession` switch

**No real Chrome.** **No real agent-run.** In-process `RunDaemon` on ephemeral
`127.0.0.1:0` with temp `BaseDir`; `OpenChromeFn` / `AgentRunProbeFn` record
calls. `SpawnFn` is injectable for `EnsureDaemon` spawn path.

| Surface | What is under test |
|---------|-------------------|
| `EnsureDaemon` | Reuse live daemon vs spawn when down (`SpawnFn` recorded) |
| `SessionNew` | POST create, OpenChrome once, no agent-run, stdout markers |
| `session new` CLI | `HandleCLI session new` end-to-end |

Depends on Phases 1–7 (`DaemonMeta`, `RunDaemon`, `POST /v1/sessions`, status).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Ensure Daemon** (`EnsureDaemon`) probes `{BaseDir}` control health at `Addr`.
When `/v1/health` is OK and `server.json` matches, return existing `DaemonMeta`
without calling `SpawnFn`. Otherwise invoke injectable `SpawnFn` (production:
detached `browser-agent serve --daemon`), poll until healthy and meta exists.

**Session New** (`SessionNew`) operator flow:

```text
EnsureDaemon(baseDir, addr)
POST /v1/sessions {session_id?}   # generate sess-<6alnum> when omitted
OpenChromeFn(sessionURL, extPath)
print stdout: session-id, export hint, inspect/interact recipes
# never AgentRunFn
```

**Operator CLI** — `browser-agent session new` parses flags (`--session-id`,
`--base-dir`, `--addr`), delegates to `SessionNew`, exit **0** on success.

**Test Client** binds loopback `:0`, uses recording hooks, and asserts HTTP
session registry state without real Chrome or agent-run.

## Decision Tree

```
browser-agent-daemon-phase8
├── ensure-daemon/                       [EnsureDaemon]
│   ├── reuse-running/                     health ok → no SpawnFn
│   └── spawn-when-down/                   down → SpawnFn called + meta
├── session-new/                         [SessionNew package API]
│   ├── create-opens-chrome/               OpenChrome once; AgentRunProbe 0
│   ├── skip-open-chrome/                    NoOpenChrome → OpenChrome 0; stdout markers
│   ├── duplicate-409/                       second same id → error / 409
│   ├── auto-generate-id/                    omit id → ^sess-[a-z0-9]{6}$
│   └── pretty-output-markers/               stdout session/export/inspect hints
└── cli-dispatch/                        [HandleCLI session new]
    └── session-new-subcommand/            CLI creates session + stdout markers
```

### Parameter significance (high → low)

1. **Entry surface** — `EnsureDaemon` vs `SessionNew` vs CLI.
2. **Daemon state** — already running vs absent (spawn path).
3. **Session id** — explicit vs auto-generated vs duplicate.
4. **Hooks** — `SpawnFn`, `OpenChromeFn`, `AgentRunProbeFn` call counts.

## Test Index

| Leaf | Scenario |
|------|----------|
| `ensure-daemon/reuse-running` | Live daemon; `EnsureDaemon` returns meta; `SpawnFn` not called |
| `ensure-daemon/spawn-when-down` | No daemon; `SpawnFn` called; health + `server.json` |
| `session-new/create-opens-chrome` | `OpenChromeFn` once with `/go` URL; `AgentRunProbeFn` never called |
| `session-new/skip-open-chrome` | `NoOpenChrome` → `OpenChromeFn` 0; session on server; stdout markers |
| `session-new/duplicate-409` | Second `SessionNew` same id → duplicate error |
| `session-new/auto-generate-id` | Omitted id → generated `sess-` + 6 alnum; session on server |
| `session-new/pretty-output-markers` | Stdout has session id, export hint, nested session recipes |
| `cli-dispatch/session-new-subcommand` | `HandleCLI session new` exit 0; session registered |

**Leaf count: 8**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase8
doctest test ./tests/browser-agent-daemon-phase8
# After implementer lands phase 8:
doctest test ./tests/browser-agent-daemon-phase7
doctest test ./tests/browser-agent-session-nested/...
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until
implementer lands phase 8). Prose contract (authoritative for GREEN):

**EnsureDaemon** — `EnsureDaemonConfig{BaseDir, Addr, SpawnFn, WaitTimeout}`;
`SpawnFn` injectable (default detached serve); returns `DaemonMeta`.

**SessionNew** — `SessionNewConfig{BaseDir, Addr, SessionID, NoOpenChrome,
OpenChromeFn, AgentRunProbeFn, Stdout, Stderr}`; when `NoOpenChrome` skip
`OpenChromeFn`; never invokes agent-run; pretty stdout.

**CLI** — `HandleCLI session new [--session-id] [--base-dir] [--addr]`; optional
`SessionNewTestHooks` for test injection when hooks are not passed via config.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

// Mode — top-level API under test.
const (
	ModeEnsureDaemon = "ensure-daemon"
	ModeSessionNew   = "session-new"
	ModeCLIDispatch  = "cli-dispatch"
)

// EnsureDaemonOp — EnsureDaemon probes.
const (
	EnsureDaemonOpReuseRunning  = "reuse-running"
	EnsureDaemonOpSpawnWhenDown = "spawn-when-down"
)

// SessionNewOp — SessionNew probes.
const (
	SessionNewOpCreateOpensChrome   = "create-opens-chrome"
	SessionNewOpSkipOpenChrome      = "skip-open-chrome"
	SessionNewOpDuplicate409        = "duplicate-409"
	SessionNewOpAutoGenerateID      = "auto-generate-id"
	SessionNewOpPrettyOutputMarkers = "pretty-output-markers"
)

// CLIDispatchOp — HandleCLI probes.
const (
	CLIDispatchOpSessionNewSubcommand = "session-new-subcommand"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	EnsureDaemonOp string
	SessionNewOp   string
	CLIDispatchOp  string

	// Session id for explicit-id leaves; empty for auto-generate.
	SessionID string

	// NoOpenChrome skips OpenChromeFn when true (session-new/skip-open-chrome).
	NoOpenChrome bool

	ReadyTimeout time.Duration
}

// Response holds ensure / session-new / CLI outcomes.
type Response struct {
	// EnsureDaemon
	DaemonMeta     browseragent.DaemonMeta
	EnsureErr      string
	SpawnFnCalled  bool
	PreExistingPID int

	// SessionNew / CLI
	SessionNewErr string
	SessionID     string
	SessionURL    string
	Stdout        string
	Stderr        string
	CLIErr        string

	OpenChromeCallCount   int
	OpenChromeSessionURL  string
	OpenChromeExtPath     string
	AgentRunProbeCallCount int

	SecondSessionNewErr string

	BaseURL string
	Addr    string

	HTTPPostStatus int
	SessionOnServer bool
	ServerSessionIDs []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeEnsureDaemon:
		return runEnsureDaemonMode(t, req)
	case ModeSessionNew:
		return runSessionNewMode(t, req)
	case ModeCLIDispatch:
		return runCLIDispatchMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runEnsureDaemonMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.EnsureDaemonOp == "" {
		t.Fatal("EnsureDaemonOp must be set by leaf Setup")
	}
	resp := &Response{}

	switch req.EnsureDaemonOp {
	case EnsureDaemonOpReuseRunning:
		srv, cleanup, err := startDaemonServer(t, req)
		if err != nil {
			return resp, err
		}
		t.Cleanup(cleanup)

		resp.BaseURL = srv.BaseURL
		resp.Addr = srv.Addr
		resp.PreExistingPID = srv.PID

		spawnCalled := false
		cfg := browseragent.EnsureDaemonConfig{
			BaseDir:     req.BaseDir,
			Addr:        srv.Addr,
			WaitTimeout: req.ReadyTimeout,
			SpawnFn: func() error {
				spawnCalled = true
				return fmt.Errorf("SpawnFn must not be called when daemon is healthy")
			},
		}
		meta, eerr := browseragent.EnsureDaemon(cfg)
		resp.SpawnFnCalled = spawnCalled
		if eerr != nil {
			resp.EnsureErr = eerr.Error()
			return resp, eerr
		}
		resp.DaemonMeta = meta
		return resp, nil

	case EnsureDaemonOpSpawnWhenDown:
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return resp, err
		}
		addr := ln.Addr().String()
		_ = ln.Close()
		if req.Addr != "" {
			addr = req.Addr
		}
		resp.Addr = addr
		resp.BaseURL = "http://" + addr

		spawnCalled := false
		var daemonCancel context.CancelFunc
		cfg := browseragent.EnsureDaemonConfig{
			BaseDir:     req.BaseDir,
			Addr:        addr,
			WaitTimeout: req.ReadyTimeout,
			SpawnFn: func() error {
				spawnCalled = true
				ctx, cancel := context.WithCancel(context.Background())
				daemonCancel = cancel
				daemonCfg := browseragent.DaemonConfig{
					Addr:    addr,
					BaseDir: req.BaseDir,
					Stdout:  io.Discard,
					Stderr:  io.Discard,
				}
				go func() {
					_, _ = browseragent.RunDaemon(ctx, daemonCfg)
				}()
				return nil
			},
		}
		meta, eerr := browseragent.EnsureDaemon(cfg)
		resp.SpawnFnCalled = spawnCalled
		if eerr != nil {
			resp.EnsureErr = eerr.Error()
			return resp, eerr
		}
		resp.DaemonMeta = meta
		if daemonCancel != nil {
			t.Cleanup(daemonCancel)
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown EnsureDaemonOp %q", req.EnsureDaemonOp)
	}
}

func runSessionNewMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionNewOp == "" {
		t.Fatal("SessionNewOp must be set by leaf Setup")
	}
	resp := &Response{}

	// Ephemeral listen addr — SessionNew/EnsureDaemon must not default to 43761
	// when another daemon may already occupy the product port.
	if strings.TrimSpace(req.Addr) == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return resp, err
		}
		req.Addr = ln.Addr().String()
		_ = ln.Close()
	}

	var openCount, agentProbeCount int
	var openURL, openExt string
	var hookMu sync.Mutex

	recordOpen := func(sessionURL, extPath string) error {
		hookMu.Lock()
		defer hookMu.Unlock()
		openCount++
		openURL = sessionURL
		openExt = extPath
		return nil
	}
	recordAgentProbe := func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error {
		hookMu.Lock()
		defer hookMu.Unlock()
		agentProbeCount++
		return nil
	}

	var stdout, stderr bytes.Buffer
	cfg := browseragent.SessionNewConfig{
		BaseDir:         req.BaseDir,
		Addr:            req.Addr,
		SessionID:       req.SessionID,
		NoOpenChrome:    req.NoOpenChrome,
		OpenChromeFn:    recordOpen,
		AgentRunProbeFn: recordAgentProbe,
		Stdout:          &stdout,
		Stderr:          &stderr,
	}

	err := browseragent.SessionNew(cfg)
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	if err != nil {
		resp.SessionNewErr = err.Error()
	}

	hookMu.Lock()
	resp.OpenChromeCallCount = openCount
	resp.OpenChromeSessionURL = openURL
	resp.OpenChromeExtPath = openExt
	resp.AgentRunProbeCallCount = agentProbeCount
	hookMu.Unlock()

	if err != nil && req.SessionNewOp != SessionNewOpDuplicate409 {
		return resp, err
	}

	// Resolve addr/base for HTTP probes after EnsureDaemon inside SessionNew.
	addr := req.Addr
	if addr == "" {
		if meta, rerr := browseragent.ReadDaemonMeta(daemonMetaPath(req.BaseDir)); rerr == nil && meta.Addr != "" {
			addr = meta.Addr
		}
	}
	if addr != "" {
		resp.Addr = addr
		resp.BaseURL = "http://" + addr
		ids, lerr := listSessionsHTTP(resp.BaseURL)
		if lerr == nil {
			resp.ServerSessionIDs = ids
		}
	}

	if req.SessionNewOp == SessionNewOpDuplicate409 {
		// First call may succeed or fail setup; leaf expects first ok then second dup.
		if resp.SessionNewErr != "" && !strings.Contains(strings.ToLower(resp.SessionNewErr), "duplicate") {
			// First call failed unexpectedly — still attempt second for RED contract.
		}
		cfg2 := cfg
		var stdout2 bytes.Buffer
		cfg2.Stdout = &stdout2
		err2 := browseragent.SessionNew(cfg2)
		resp.SecondSessionNewErr = ""
		if err2 != nil {
			resp.SecondSessionNewErr = err2.Error()
		}
		resp.Stdout = stdout2.String()
		return resp, nil
	}

	if err != nil {
		return resp, err
	}

	sid := req.SessionID
	if sid == "" {
		sid = extractSessionIDFromStdout(resp.Stdout)
	}
	resp.SessionID = sid
	if sid != "" && resp.BaseURL != "" {
		resp.SessionOnServer = sessionIDInList(resp.ServerSessionIDs, sid)
	}
	return resp, nil
}

func runCLIDispatchMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLIDispatchOp == "" {
		t.Fatal("CLIDispatchOp must be set by leaf Setup")
	}

	var openCount int
	var openURL, openExt string
	var hookMu sync.Mutex

	hooks := &browseragent.SessionNewTestHooks{
		OpenChromeFn: func(sessionURL, extPath string) error {
			hookMu.Lock()
			defer hookMu.Unlock()
			openCount++
			openURL = sessionURL
			openExt = extPath
			return nil
		},
		AgentRunProbeFn: func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error {
			hookMu.Lock()
			defer hookMu.Unlock()
			return fmt.Errorf("agent-run must not be invoked during session new")
		},
	}
	inj.SessionNewTestHooks = hooks
	defer func() { inj.SessionNewTestHooks = nil }()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	resp := &Response{Addr: addr}

	var stdout, stderr bytes.Buffer
	host, portStr, _ := net.SplitHostPort(addr)
	args := []string{
		"session", "new",
		"--base-dir", req.BaseDir,
		"--host", host,
		"--server-port", portStr,
	}
	if req.SessionID != "" {
		args = append(args, "--session-id", req.SessionID)
	}
	cliErr := browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
	}

	hookMu.Lock()
	resp.OpenChromeCallCount = openCount
	resp.OpenChromeSessionURL = openURL
	resp.OpenChromeExtPath = openExt
	hookMu.Unlock()

	resp.BaseURL = "http://" + addr
	ids, lerr := listSessionsHTTP(resp.BaseURL)
	if lerr == nil {
		resp.ServerSessionIDs = ids
	}

	sid := req.SessionID
	if sid == "" {
		sid = extractSessionIDFromStdout(resp.Stdout)
	}
	resp.SessionID = sid
	if sid != "" {
		resp.SessionOnServer = sessionIDInList(resp.ServerSessionIDs, sid)
	}

	return resp, cliErr
}

// --- harness helpers ---

type daemonServer struct {
	BaseURL string
	Addr    string
	PID     int
	cancel  context.CancelFunc
	done    <-chan error
}

func startDaemonServer(t *testing.T, req *Request) (*daemonServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by root Setup")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	if req.Addr != "" {
		addr = req.Addr
	}

	ready := req.ReadyTimeout
	if ready <= 0 {
		ready = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	cfg := browseragent.DaemonConfig{
		Addr:    addr,
		BaseDir: req.BaseDir,
		Stdout:  io.Discard,
		Stderr:  io.Discard,
	}

	done := make(chan error, 1)
	go func() {
		_, err := browseragent.RunDaemon(ctx, cfg)
		done <- err
	}()

	baseURL := "http://" + addr
	if err := waitHealth(baseURL, ready); err != nil {
		cancel()
		<-done
		return nil, nil, fmt.Errorf("RunDaemon never healthy at %s: %w", baseURL, err)
	}

	meta, _ := browseragent.ReadDaemonMeta(daemonMetaPath(req.BaseDir))
	srv := &daemonServer{
		BaseURL: baseURL,
		Addr:    addr,
		PID:     meta.PID,
		cancel:  cancel,
		done:    done,
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
		if healthOK(baseURL) {
			return nil
		}
		last = fmt.Errorf("health not ok")
		time.Sleep(20 * time.Millisecond)
	}
	return last
}

func healthOK(baseURL string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/health", nil)
	if err != nil {
		return false
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	return res.StatusCode == http.StatusOK
}

func listSessionsHTTP(baseURL string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/sessions", nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /v1/sessions status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err != nil {
		var wrap map[string]any
		if err2 := json.Unmarshal(body, &wrap); err2 != nil {
			return nil, fmt.Errorf("parse sessions: %w", err)
		}
		items, _ := wrap["sessions"].([]any)
		for _, it := range items {
			if m, ok := it.(map[string]any); ok {
				arr = append(arr, m)
			}
		}
	}
	ids := make([]string, 0, len(arr))
	for _, item := range arr {
		if id, ok := item["session_id"].(string); ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func daemonMetaPath(baseDir string) string {
	return filepath.Join(baseDir, "server.json")
}

func sessionIDInList(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

var sessIDLineRe = regexp.MustCompile(`(?i)(?:session[_ -]?id|session-id)\s*[:=]\s*([a-zA-Z0-9][a-zA-Z0-9._-]{0,63})`)
var sessGenerateRe = regexp.MustCompile(`^sess-[a-z0-9]{6}$`)

func extractSessionIDFromStdout(stdout string) string {
	for _, line := range strings.Split(stdout, "\n") {
		if m := sessIDLineRe.FindStringSubmatch(line); len(m) > 1 {
			return m[1]
		}
	}
	return ""
}

var (
	_ = sync.Mutex{}
	_ = os.Getpid
)
```