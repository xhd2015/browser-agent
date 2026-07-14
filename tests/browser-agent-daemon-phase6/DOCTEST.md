# browser-agent daemon Phase 6 — graceful shutdown + kill-existing

Phase 6 adds **`POST /v1/shutdown`** (202 Accepted), client helpers
**`ShutdownDaemon`** / **`KillExistingDaemon`**, and CLI **`serve --kill-existing`**.

**No real Chrome.** **No real agent-run.** In-process `RunDaemon` on ephemeral
`127.0.0.1:0` with temp `BaseDir`; shutdown via HTTP self-request (not ctx cancel
for shutdown leaves).

| Surface | What is under test |
|---------|-------------------|
| `POST /v1/shutdown` | Returns **202**; triggers graceful drain |
| `RunDaemon` + shutdown | Loop exits; `server.json` removed |
| `KillExistingDaemon` | Reads meta, graceful shutdown, force-kill if needed |
| `serve --kill-existing` | Kills existing daemon, pretty stderr, then serve |

Depends on Phase 5 (`RunDaemon`, `DaemonMeta`, `server.json`).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Host** (`RunDaemon`) binds the control HTTP server, writes
`{BaseDir}/server.json`, and blocks until shutdown. An injectable
**ShutdownGracePeriod** can delay drain (for force-kill tests).

**Shutdown Endpoint** (`POST /v1/shutdown`) accepts a graceful shutdown request,
returns **202 Accepted**, and signals the daemon host to drain and exit.

**Client Helpers** — **`ShutdownDaemon(baseURL, timeout)`** POSTs shutdown and
polls `/v1/health` until down; **`KillExistingDaemon(baseDir, timeout)`** reads
`server.json`, POSTs shutdown, waits up to timeout (default 10s), then **SIGKILL**
the recorded pid if still alive and removes stale `server.json`.

**Operator CLI** — `browser-agent serve --kill-existing` calls
`KillExistingDaemon` before blocking serve and prints operator-facing status on
stderr.

**Test Client** starts `RunDaemon` in a goroutine, probes HTTP routes, invokes
helpers / CLI, and asserts meta removal and health-down without real Chrome.

```text
RunDaemon -> server.json + GET /v1/health OK
POST /v1/shutdown -> 202 -> drain -> RunDaemon exits -> server.json gone

KillExistingDaemon(baseDir) -> read server.json -> POST shutdown -> wait
  -> (timeout) SIGKILL pid -> remove server.json

HandleCLI serve --kill-existing -> KillExistingDaemon -> RunDaemon (new)
```

## Decision Tree

```
browser-agent-daemon-phase6
├── v1-shutdown/                         [POST /v1/shutdown HTTP]
│   ├── post-202/                          POST returns 202 Accepted
│   └── server-stops/                      POST shutdown; RunDaemon exits; meta gone
├── kill-existing/                       [KillExistingDaemon helper]
│   ├── graceful-within-timeout/           graceful shutdown; pid down; meta gone
│   └── force-after-timeout/               slow grace + short timeout → force kill
└── cli-kill/                            [HandleCLI serve --kill-existing]
    └── kill-existing-ok/                  kills prior daemon; pretty stderr; exit 0
```

### Parameter significance (high → low)

1. **Entry surface** — HTTP shutdown vs `KillExistingDaemon` vs CLI.
2. **Shutdown outcome** — 202 only vs daemon exit vs meta removal vs force kill.
3. **Timing** — default grace vs injected `ShutdownGracePeriod` + short kill timeout.

## Test Index

| Leaf | Scenario |
|------|----------|
| `v1-shutdown/post-202` | `POST /v1/shutdown` → **202** |
| `v1-shutdown/server-stops` | Shutdown POST; `RunDaemon` exits; `server.json` absent |
| `kill-existing/graceful-within-timeout` | `KillExistingDaemon` succeeds; health down; meta gone |
| `kill-existing/force-after-timeout` | Slow grace + short timeout → force path; meta gone |
| `cli-kill/kill-existing-ok` | `serve --kill-existing` kills prior daemon; stderr status; exit 0 |

**Leaf count: 5**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase6
doctest test ./tests/browser-agent-daemon-phase6
# After implementer lands phase 6:
doctest test ./tests/browser-agent-daemon-phase5
doctest test ./tests/browser-agent/...
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until
implementer lands phase 6):

- `POST /v1/shutdown` on registry control handler → **202**
- `ShutdownDaemon(baseURL string, timeout time.Duration) error`
- `KillExistingDaemon(baseDir string, timeout time.Duration) error`
- `DaemonConfig.ShutdownGracePeriod time.Duration` (or package test hook) — delay drain
- `HandleCLI serve --kill-existing` — calls `KillExistingDaemon`, pretty stderr

```go
import (
	"bytes"
	"context"
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
	ModeV1Shutdown   = "v1-shutdown"
	ModeKillExisting = "kill-existing"
	ModeCLIKill      = "cli-kill"
)

// V1ShutdownOp — HTTP shutdown probes.
const (
	V1ShutdownOpPost202     = "post-202"
	V1ShutdownOpServerStops = "server-stops"
)

// KillExistingOp — KillExistingDaemon probes.
const (
	KillExistingOpGraceful = "graceful-within-timeout"
	KillExistingOpForce    = "force-after-timeout"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	V1ShutdownOp   string
	KillExistingOp string

	// ShutdownGracePeriod injects slow drain on RunDaemon (force-kill leaf).
	ShutdownGracePeriod time.Duration
	// KillTimeout passed to KillExistingDaemon (force leaf uses short value).
	KillTimeout time.Duration

	ReadyTimeout time.Duration
	ShutdownWait time.Duration
}

// Response holds shutdown / kill / CLI outcomes.
type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string

	BaseURL string
	Addr    string

	DaemonMetaPath   string
	DaemonMetaExists bool
	DaemonMeta       browseragent.DaemonMeta
	ReadMetaErr      string

	DaemonExited bool
	HealthDown   bool

	KillErr string

	Stderr string
	Stdout string

	KillMessageSeen bool
	ExitCode        int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeV1Shutdown:
		return runV1ShutdownMode(t, req)
	case ModeKillExisting:
		return runKillExistingMode(t, req)
	case ModeCLIKill:
		return runCLIKillMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runV1ShutdownMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.V1ShutdownOp == "" {
		t.Fatal("V1ShutdownOp must be set by leaf Setup")
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

	switch req.V1ShutdownOp {
	case V1ShutdownOpPost202:
		status, ct, body, err := doPOST(srv.BaseURL+"/v1/shutdown", nil)
		if err != nil {
			return resp, err
		}
		resp.StatusCode = status
		resp.ContentType = ct
		resp.Body = body
		resp.BodyString = string(body)
		return resp, nil

	case V1ShutdownOpServerStops:
		metaPath := daemonMetaPath(req.BaseDir)
		if _, err := os.Stat(metaPath); err != nil {
			return resp, fmt.Errorf("server.json missing before shutdown: %w", err)
		}
		resp.DaemonMetaExists = true

		status, _, body, err := doPOST(srv.BaseURL+"/v1/shutdown", nil)
		if err != nil {
			return resp, err
		}
		resp.StatusCode = status
		resp.BodyString = string(body)

		wait := req.ShutdownWait
		if wait <= 0 {
			wait = 5 * time.Second
		}
		select {
		case <-srv.done:
			resp.DaemonExited = true
		case <-time.After(wait):
			return resp, fmt.Errorf("RunDaemon did not exit within %v after POST /v1/shutdown", wait)
		}
		if err := waitHealthDown(srv.BaseURL, wait); err != nil {
			return resp, err
		}
		resp.HealthDown = true

		_, err = os.Stat(metaPath)
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
		return nil, fmt.Errorf("unknown V1ShutdownOp %q", req.V1ShutdownOp)
	}
}

func runKillExistingMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.KillExistingOp == "" {
		t.Fatal("KillExistingOp must be set by leaf Setup")
	}
	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	// Do not defer cleanup — KillExistingDaemon must stop the daemon.

	resp := &Response{
		BaseURL:        srv.BaseURL,
		Addr:           srv.Addr,
		DaemonMetaPath: daemonMetaPath(req.BaseDir),
	}

	meta, _, exists, readErr := readDaemonMetaFile(req.BaseDir)
	resp.DaemonMeta = meta
	resp.DaemonMetaExists = exists
	if readErr != nil {
		resp.ReadMetaErr = readErr.Error()
	}

	timeout := req.KillTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	killErr := browseragent.KillExistingDaemon(req.BaseDir, timeout)
	if killErr != nil {
		resp.KillErr = killErr.Error()
		cleanup()
		return resp, killErr
	}

	wait := req.ShutdownWait
	if wait <= 0 {
		wait = 5 * time.Second
	}
	if err := waitHealthDown(srv.BaseURL, wait); err != nil {
		cleanup()
		return resp, err
	}
	resp.HealthDown = true

	select {
	case <-srv.done:
		resp.DaemonExited = true
	case <-time.After(wait):
	}

	_, err = os.Stat(daemonMetaPath(req.BaseDir))
	if err == nil {
		resp.DaemonMetaExists = true
		cleanup()
		return resp, nil
	}
	if os.IsNotExist(err) {
		resp.DaemonMetaExists = false
		return resp, nil
	}
	cleanup()
	return resp, err
}

func runCLIKillMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	// First daemon must be killed by CLI; defer only as backstop.
	defer cleanup()

	firstBaseURL := srv.BaseURL
	if err := waitHealth(firstBaseURL, req.ReadyTimeout); err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	args := []string{
		"serve",
		"--kill-existing",
		"--no-open-chrome",
		"--no-agent-run",
		"--addr", srv.Addr,
		"--base-dir", req.BaseDir,
	}
	done := make(chan error, 1)
	go func() {
		done <- browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)
	}()

	// Wait until first daemon is down (killed by --kill-existing).
	downWait := req.ShutdownWait
	if downWait <= 0 {
		downWait = 8 * time.Second
	}
	if err := waitHealthDown(firstBaseURL, downWait); err != nil {
		return &Response{
			BaseURL: firstBaseURL,
			Addr:    srv.Addr,
			Stderr:  stderr.String(),
			Stdout:  stdout.String(),
		}, fmt.Errorf("first daemon still healthy after --kill-existing: %w", err)
	}

	// Poll stderr for operator-facing kill status.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		low := strings.ToLower(stderr.String())
		if strings.Contains(low, "kill") ||
			strings.Contains(low, "shutdown") ||
			strings.Contains(low, "existing") ||
			strings.Contains(low, "stopped") {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	resp := &Response{
		BaseURL: firstBaseURL,
		Addr:    srv.Addr,
		Stderr:  stderr.String(),
		Stdout:  stdout.String(),
	}
	low := strings.ToLower(resp.Stderr)
	resp.KillMessageSeen = strings.Contains(low, "kill") ||
		strings.Contains(low, "shutdown") ||
		strings.Contains(low, "existing") ||
		strings.Contains(low, "stopped")

	// Second serve (from HandleCLI) should be up on same addr; shut it down for exit 0.
	secondBaseURL := "http://" + srv.Addr
	if err := waitHealth(secondBaseURL, 5*time.Second); err != nil {
		return resp, fmt.Errorf("new serve never healthy after --kill-existing: %w", err)
	}
	if err := browseragent.ShutdownDaemon(secondBaseURL, 5*time.Second); err != nil {
		return resp, fmt.Errorf("ShutdownDaemon second serve: %w", err)
	}

	select {
	case err := <-done:
		if err != nil {
			resp.KillErr = err.Error()
			return resp, err
		}
		resp.ExitCode = 0
		return resp, nil
	case <-time.After(downWait):
		return resp, fmt.Errorf("HandleCLI did not return after shutting down second serve")
	}
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
	if req.ShutdownGracePeriod > 0 {
		cfg.ShutdownGracePeriod = req.ShutdownGracePeriod
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

func waitHealthDown(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/health", nil)
		if err != nil {
			cancel()
			time.Sleep(20 * time.Millisecond)
			continue
		}
		res, err := http.DefaultClient.Do(req)
		cancel()
		if err != nil {
			return nil
		}
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("health still up at %s after %v", baseURL, timeout)
}

func doPOST(rawURL string, body []byte) (int, string, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, r)
	if err != nil {
		return 0, "", nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer res.Body.Close()
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, res.Header.Get("Content-Type"), nil, err
	}
	return res.StatusCode, res.Header.Get("Content-Type"), out, nil
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