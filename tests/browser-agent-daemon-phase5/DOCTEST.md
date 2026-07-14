# browser-agent daemon Phase 5 — blocking serve refactor

Splits **daemon host** (`RunDaemon`) from **session create** (`Run` with
`SessionID`). Daemon starts an empty registry, writes `server.json`, blocks
until ctx cancel. `browser-agent serve` uses `RunDaemon`. `serve --session-id`
emits a deprecation warning on stderr but remains backward compatible.

**No real Chrome.** **No real agent-run.** Integration via in-process
`RunDaemon` / `Run` with ephemeral `127.0.0.1:0` listen and temp `BaseDir`.

| Surface | What is under test |
|---------|-------------------|
| `RunDaemon` | Empty registry; `/v1/health` 200; `/v1/sessions` → `[]`; writes/removes `server.json` |
| `Run` compat | `Config.SessionID` still creates one session and serves (bridge for `tests/browser-agent/`) |
| `serve --session-id` | `HandleCLI` prints deprecation warning on stderr before blocking serve |

Depends on Phases 1–4 (`DaemonMeta`, `SessionRegistry`, registry HTTP routes).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Host** (`RunDaemon`) starts a **SessionRegistry** with **zero**
pre-created sessions, binds the control HTTP server, writes **daemon discovery**
`{BaseDir}/server.json` (`DaemonMeta`: pid, addr, base_url, base_dir,
started_at), and **blocks** until the caller cancels `ctx`. On clean shutdown it
**removes** `server.json`. No Chrome launch, no agent-run, no auto-session.

**Session Serve** (`Run` with `Config.SessionID`) remains **backward compat**:
creates exactly one registry session then behaves like today's single-session
serve (used by existing `tests/browser-agent/` harness).

**Operator CLI** `browser-agent serve` calls **`RunDaemon`** (blocking).
`serve --session-id <id>` prints a **deprecation warning** on stderr; compat path
may delegate to `Run` with that session or equivalent create-on-start behavior.

**Test Client** binds loopback `:0`, starts `RunDaemon` or `Run` in a
goroutine, probes HTTP routes, reads/removes `server.json`, cancels ctx for
shutdown assertions.

```text
RunDaemon(ctx, DaemonConfig) -> empty registry + server.json + block
ctx cancel -> RemoveDaemonMeta(server.json)

Run(ctx, Config{SessionID}) -> registry.Create(SessionID) -> serve (compat)

HandleCLI serve --session-id -> stderr deprecation warning -> serve
```

## Decision Tree

```
browser-agent-daemon-phase5
├── run-daemon/                          [RunDaemon blocking host]
│   ├── starts-health-ok/                  GET /v1/health → 200
│   ├── zero-sessions/                     GET /v1/sessions → []
│   ├── writes-server-json/                server.json pid/addr after start
│   └── shutdown-removes-json/             cancel ctx → server.json gone
├── run-compat/                          [Run(SessionID) regression bridge]
│   └── single-session-still-works/        Run + GET /v1/session?session=id
└── cli-serve/                           [HandleCLI serve flags]
    └── deprecated-session-id-warn/        serve --session-id → stderr deprecation
```

### Parameter significance (high → low)

1. **Entry API** — `RunDaemon` vs `Run` compat vs CLI `serve`.
2. **Daemon outcome** — health vs empty list vs meta write vs meta remove on shutdown.
3. **Compat / CLI details** — session id probe; deprecation stderr tokens.

## Test Index

| Leaf | Scenario |
|------|----------|
| `run-daemon/starts-health-ok` | `RunDaemon`; `GET /v1/health` → 200 |
| `run-daemon/zero-sessions` | `RunDaemon`; `GET /v1/sessions` → `[]` |
| `run-daemon/writes-server-json` | After start, `server.json` has pid, addr, base_url, base_dir |
| `run-daemon/shutdown-removes-json` | Cancel ctx; `server.json` absent |
| `run-compat/single-session-still-works` | `Run` with `SessionID`; `GET /v1/session?session=` → 200 |
| `cli-serve/deprecated-session-id-warn` | `HandleCLI serve --session-id` stderr mentions deprecation |

**Leaf count: 6**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase5
doctest test ./tests/browser-agent-daemon-phase5
# After implementer lands phase 5:
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-serve-runtime/...
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until
implementer lands phase 5):

- `RunDaemon(ctx context.Context, cfg DaemonConfig) (*Result, error)` — empty
  registry; writes/removes `{BaseDir}/server.json`; blocks until ctx cancel
- `DaemonConfig{Addr, BaseDir, Stdout, Stderr}` — no `SessionID`
- `Run(ctx, Config)` unchanged compat when `SessionID` set
- `HandleCLI serve` uses `RunDaemon`; `serve --session-id` deprecation on stderr

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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level API under test.
const (
	ModeRunDaemon = "run-daemon"
	ModeRunCompat = "run-compat"
	ModeCLIServe  = "cli-serve"
)

// RunDaemonOp — daemon host probes.
const (
	RunDaemonOpHealthOK         = "health-ok"
	RunDaemonOpZeroSessions     = "zero-sessions"
	RunDaemonOpWritesServerJSON = "writes-server-json"
	RunDaemonOpShutdownRemoves  = "shutdown-removes-json"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	RunDaemonOp string

	// Run compat
	SessionID    string
	NoOpenChrome bool
	NoAgentRun   bool

	// CLI serve
	CLISessionID string

	ReadyTimeout time.Duration
	ShutdownWait time.Duration
}

// Response holds daemon / HTTP / CLI outcomes.
type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string
	Raw         map[string]any

	BaseURL string
	Addr    string

	SessionsListIDs []string

	DaemonMeta       browseragent.DaemonMeta
	DaemonMetaPath   string
	DaemonMetaExists bool
	DaemonMetaRaw    []byte
	ReadMetaErr      string

	Stderr string
	Stdout string

	DeprecationWarnSeen bool

	SessionIDField       string
	ExtensionConnected   bool
	ProbeURL             string

	RunErrText string
	ExitCode   int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeRunDaemon:
		return runDaemonMode(t, req)
	case ModeRunCompat:
		return runCompatMode(t, req)
	case ModeCLIServe:
		return runCLIServeMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runDaemonMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.RunDaemonOp == "" {
		t.Fatal("RunDaemonOp must be set by leaf Setup")
	}
	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		BaseURL:        srv.BaseURL,
		Addr:           srv.Addr,
		DaemonMetaPath: daemonMetaPath(req.BaseDir),
	}

	switch req.RunDaemonOp {
	case RunDaemonOpHealthOK:
		status, ct, body, err := doGET(srv.BaseURL + "/v1/health")
		if err != nil {
			return resp, err
		}
		resp.StatusCode = status
		resp.ContentType = ct
		resp.Body = body
		resp.BodyString = string(body)
		return resp, nil

	case RunDaemonOpZeroSessions:
		status, ct, body, err := doGET(srv.BaseURL + "/v1/sessions")
		if err != nil {
			return resp, err
		}
		resp.StatusCode = status
		resp.ContentType = ct
		resp.Body = body
		resp.BodyString = string(body)
		parseSessionsListJSON(resp, body)
		return resp, nil

	case RunDaemonOpWritesServerJSON:
		meta, raw, exists, readErr := readDaemonMetaFile(req.BaseDir)
		resp.DaemonMeta = meta
		resp.DaemonMetaRaw = raw
		resp.DaemonMetaExists = exists
		if readErr != nil {
			resp.ReadMetaErr = readErr.Error()
		}
		return resp, nil

	case RunDaemonOpShutdownRemoves:
		metaPath := daemonMetaPath(req.BaseDir)
		if _, err := os.Stat(metaPath); err != nil {
			return resp, fmt.Errorf("server.json missing before shutdown: %w", err)
		}
		resp.DaemonMetaExists = true
		srv.cancel()
		wait := req.ShutdownWait
		if wait <= 0 {
			wait = 3 * time.Second
		}
		select {
		case <-srv.done:
		case <-time.After(wait):
			return resp, fmt.Errorf("RunDaemon did not exit within %v after cancel", wait)
		}
		_, err := os.Stat(metaPath)
		if err == nil {
			resp.DaemonMetaExists = true
			return resp, nil
		}
		if os.IsNotExist(err) {
			resp.DaemonMetaExists = false
			return resp, nil
		}
		return resp, err

	default:
		return nil, fmt.Errorf("unknown RunDaemonOp %q", req.RunDaemonOp)
	}
}

func runCompatMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionID == "" {
		t.Fatal("SessionID must be set for run-compat")
	}
	srv, cleanup, err := startAgentRunServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		BaseURL: srv.BaseURL,
		Addr:    srv.Addr,
	}
	if err := fillSessionProbe(resp, srv.BaseURL, req.SessionID); err != nil {
		return resp, err
	}
	return resp, nil
}

func runCLIServeMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLISessionID == "" {
		t.Fatal("CLISessionID must be set for cli-serve")
	}
	if req.Addr == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}
		req.Addr = ln.Addr().String()
		_ = ln.Close()
	}

	var stdout, stderr bytes.Buffer
	args := []string{
		"serve",
		"--session-id", req.CLISessionID,
		"--no-open-chrome",
		"--no-agent-run",
		"--addr", req.Addr,
		"--base-dir", req.BaseDir,
	}
	done := make(chan error, 1)
	go func() {
		done <- browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)
	}()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		s := stderr.String()
		if strings.Contains(strings.ToLower(s), "deprecat") {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Addr:   req.Addr,
	}
	low := strings.ToLower(resp.Stderr)
	resp.DeprecationWarnSeen = strings.Contains(low, "deprecat")
	return resp, nil
}

// --- RunDaemon harness ---

type daemonServer struct {
	BaseURL string
	Addr    string
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

	srv := &daemonServer{
		BaseURL: baseURL,
		Addr:    addr,
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

// --- Run() compat harness ---

type agentRunServer struct {
	BaseURL string
	Addr    string
	cancel  context.CancelFunc
	done    <-chan error
}

func startAgentRunServer(t *testing.T, req *Request) (*agentRunServer, func(), error) {
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

	ready := req.ReadyTimeout
	if ready <= 0 {
		ready = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      req.BaseDir,
		SessionID:    req.SessionID,
		NoOpenChrome: true,
		NoAgentRun:   true,
		Stdout:       io.Discard,
		Stderr:       io.Discard,
	}
	if req.NoOpenChrome {
		cfg.NoOpenChrome = true
	}
	if req.NoAgentRun {
		cfg.NoAgentRun = true
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
		return nil, nil, fmt.Errorf("Run never healthy at %s: %w", baseURL, err)
	}

	srv := &agentRunServer{
		BaseURL: baseURL,
		Addr:    addr,
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

func doGET(rawURL string) (int, string, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, "", nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, res.Header.Get("Content-Type"), nil, err
	}
	return res.StatusCode, res.Header.Get("Content-Type"), body, nil
}

func fillSessionProbe(resp *Response, baseURL, sessionID string) error {
	u := baseURL + "/v1/session?session=" + sessionID
	status, ct, body, err := doGET(u)
	if err != nil {
		return err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = body
	resp.BodyString = string(body)
	resp.ProbeURL = u
	parseSessionJSON(resp, body)
	return nil
}

func parseSessionsListJSON(resp *Response, body []byte) {
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err != nil {
		var wrap map[string]any
		if err2 := json.Unmarshal(body, &wrap); err2 == nil {
			if items, ok := wrap["sessions"].([]any); ok {
				for _, it := range items {
					if m, ok := it.(map[string]any); ok {
						arr = append(arr, m)
					}
				}
			}
		}
	}
	resp.Raw = map[string]any{"sessions": arr}
	ids := make([]string, 0, len(arr))
	for _, item := range arr {
		if id, ok := item["session_id"].(string); ok {
			ids = append(ids, id)
		}
	}
	resp.SessionsListIDs = ids
}

func parseSessionJSON(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.Raw = raw
	resp.SessionIDField, _ = raw["session_id"].(string)
	if ext, ok := raw["extension"].(map[string]any); ok {
		resp.ExtensionConnected, _ = ext["connected"].(bool)
	}
}

func daemonMetaPath(baseDir string) string {
	return filepath.Join(baseDir, "server.json")
}

func readDaemonMetaFile(baseDir string) (browseragent.DaemonMeta, []byte, bool, error) {
	path := daemonMetaPath(baseDir)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return browseragent.DaemonMeta{}, nil, false, err
		}
		return browseragent.DaemonMeta{}, nil, false, err
	}
	meta, err := browseragent.ReadDaemonMeta(path)
	if err != nil {
		return browseragent.DaemonMeta{}, raw, true, err
	}
	return meta, raw, true, nil
}

var (
	_ = sync.Mutex{}
	_ = io.Discard
)
```