# browser-agent serve --stop + less-flags + cli-color

Greenfield feature: **`serve --stop`**, less-flags migration for `cliServe`, unknown-flag
errors, mode-flag mutual exclusion, and cli-color on serve operator stderr.

**No real Chrome.** **No real agent-run.** In-process `RunDaemon` on ephemeral
`127.0.0.1:0` with temp `BaseDir`; stop/kill via `KillExistingDaemon` / HTTP
shutdown helpers.

| Surface | What is under test |
|---------|-------------------|
| `serve --stop` | Kill existing daemon only; no second `RunDaemon` bind |
| Mode flags | `--stop`, `--status`, `--kill-existing` mutually exclusive |
| less-flags | Unknown serve flags → fatal parse error |
| `serve --help` | Serve-specific help documents `--stop`, `--color`, `--no-color` |
| cli-color | Force-on ANSI, `--no-color`, `NO_COLOR=1`, flag conflict |
| Regression | `serve --kill-existing` still kills prior daemon and restarts |

Depends on Phase 5–6 (`RunDaemon`, `DaemonMeta`, `KillExistingDaemon`,
`serve --kill-existing`, `ShutdownDaemon`).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator CLI serve** (`browseragent/cli.go` via less-flags):

```text
cliServe — parse serve flags (less-flags); mode flags --stop | --status | --kill-existing
           are mutually exclusive (count > 1 → fatal error, exit 1)
           --stop → KillExistingDaemon(baseDir) only; no RunDaemon; operator stderr
           --kill-existing → KillExistingDaemon then RunDaemon (existing phase 6 path)
           --status → read-only JSON probe (never colorized)
           unknown flags → less-flags error (exit 1)
           serve --help → serve-specific usage (not fullHelp)
           cli-color on operator stderr (--stop, --kill-existing, warnings, errors)
```

**Daemon Host** (`RunDaemon`) binds control HTTP, writes `{BaseDir}/server.json`, blocks
until shutdown. Harness starts daemon on ephemeral addr for stop/regression leaves.

**Client Helpers** — `KillExistingDaemon(baseDir, timeout)` reads `server.json`, POSTs
shutdown, waits, force-kills if needed, removes stale meta.

**Test Client** starts `RunDaemon` in a goroutine when needed, invokes `HandleCLI` with
explicit env map (no ambient `NO_COLOR` leakage), and asserts health/meta/CLI outcomes.

```text
RunDaemon -> server.json + GET /v1/health OK
HandleCLI serve --stop -> KillExistingDaemon -> health down -> meta gone -> exit 0 (no re-bind)

HandleCLI serve --stop (no meta) -> warning on stderr -> exit 0

HandleCLI serve --color --stop -> colored operator stderr on pipe when --color forced
```

## Decision Tree

```
browser-agent-serve-stop
├── stop/                              [HandleCLI serve --stop]
│   ├── running-daemon/                  kills daemon; health down; meta gone; exit 0; no re-bind
│   └── not-running/                     exit 0; stderr warning: no daemon running
├── mode-conflict/                     [mutually exclusive mode flags]
│   ├── stop-with-status/
│   └── stop-with-kill-existing/
├── unknown-flag/                      [less-flags errors]
│   └── serve-typo/
├── help/                              [serve-level --help]
│   └── documents-stop-and-color/
├── color/                             [cli-color on serve stderr]
│   ├── force-on/                        --color on pipe → ANSI in stderr
│   ├── no-color-flag/                   --no-color → no ANSI
│   ├── no-color-env/                    NO_COLOR=1 auto → no ANSI
│   └── color-conflict/                  --color --no-color → exit 1
└── regression/                        [existing behavior preserved]
    └── kill-existing-still-works/         serve --kill-existing still kills + starts (phase6 parity)
```

### Parameter significance (high → low)

1. **Entry surface** — stop vs conflict vs parse error vs help vs color vs regression.
2. **Daemon state** — running vs absent (stop leaves).
3. **Color input** — force `--color`, `--no-color`, `NO_COLOR` env, or conflict.

## Test Index

| Leaf | Scenario |
|------|----------|
| `stop/running-daemon` | `--stop` kills running daemon; meta gone; no second bind; exit 0 |
| `stop/not-running` | `--stop` with no meta → exit 0 + `warning: no daemon running` |
| `mode-conflict/stop-with-status` | `--stop --status` → exit 1; mutually exclusive error |
| `mode-conflict/stop-with-kill-existing` | `--stop --kill-existing` → exit 1; mutually exclusive error |
| `unknown-flag/serve-typo` | `serve --foo` → exit 1; unrecognized flag |
| `help/documents-stop-and-color` | `serve --help` documents `--stop`, `--color`, `--no-color` |
| `color/force-on` | `serve --color --stop` on pipe → ANSI in stderr |
| `color/no-color-flag` | `serve --no-color --stop` → no ANSI in stderr |
| `color/no-color-env` | `NO_COLOR=1` without color flags → no ANSI in stderr |
| `color/color-conflict` | `--color --no-color` → exit 1 |
| `regression/kill-existing-still-works` | `serve --kill-existing` kills prior daemon; stderr status; exit 0 |

**Leaf count: 11**

## How to Run

```sh
doctest vet ./tests/browser-agent-serve-stop
doctest test ./tests/browser-agent-serve-stop          # RED after design
doctest test ./tests/browser-agent-daemon-phase6/...     # regression
doctest test ./tests/browser-agent-daemon-phase10/...    # help updates (implementer)
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until implementer
lands serve-stop feature):

- `serve --stop` via less-flags + `KillExistingDaemon` (no `RunDaemon`)
- Mode-flag mutual exclusion (`--stop`, `--status`, `--kill-existing`)
- Unknown serve flags via less-flags
- Serve-specific `--help` with `--stop`, `--color`, `--no-color`
- cli-color on serve operator stderr
- Shared serve path for `HandleCLI` and `main.go`

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

// Mode — top-level surface under test.
const (
	ModeStop         = "stop"
	ModeModeConflict = "mode-conflict"
	ModeUnknownFlag  = "unknown-flag"
	ModeHelp         = "help"
	ModeColor        = "color"
	ModeRegression   = "regression"
)

// StopOp — serve --stop probes.
const (
	StopOpRunningDaemon = "running-daemon"
	StopOpNotRunning    = "not-running"
)

// ModeConflictOp — mutually exclusive mode flag pairs.
const (
	ModeConflictOpStopStatus      = "stop-with-status"
	ModeConflictOpStopKillExisting = "stop-with-kill-existing"
)

// ColorOp — cli-color probes on serve stderr.
const (
	ColorOpForceOn       = "force-on"
	ColorOpNoColorFlag   = "no-color-flag"
	ColorOpNoColorEnv    = "no-color-env"
	ColorOpColorConflict = "color-conflict"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	StopOp         string
	ModeConflictOp string
	ColorOp        string

	CLIArgs []string
	Env     map[string]string

	ReadyTimeout time.Duration
	ShutdownWait time.Duration
	KillTimeout  time.Duration
}

// Response holds serve-stop / color / help / regression outcomes.
type Response struct {
	BaseURL string
	Addr    string

	DaemonMetaPath   string
	DaemonMetaExists bool

	DaemonExited bool
	HealthDown   bool

	Stderr string
	Stdout string

	CLIErr   string
	ExitCode int

	HelpText string

	StopMessageSeen  bool
	WarningSeen      bool
	KillMessageSeen  bool
	HasANSI          bool
	MutualExclSeen   bool
	UnknownFlagSeen  bool
	ColorConflictSeen bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeStop:
		return runStopMode(t, req)
	case ModeModeConflict:
		return runModeConflictMode(t, req)
	case ModeUnknownFlag:
		return runUnknownFlagMode(t, req)
	case ModeHelp:
		return runHelpMode(t, req)
	case ModeColor:
		return runColorMode(t, req)
	case ModeRegression:
		return runRegressionMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runStopMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.StopOp == "" {
		t.Fatal("StopOp must be set by leaf Setup")
	}

	resp := &Response{
		DaemonMetaPath: daemonMetaPath(req.BaseDir),
	}

	switch req.StopOp {
	case StopOpRunningDaemon:
		srv, cleanup, err := startDaemonServer(t, req)
		if err != nil {
			return resp, err
		}
		defer cleanup()

		resp.BaseURL = srv.BaseURL
		resp.Addr = srv.Addr

		if err := waitHealth(srv.BaseURL, req.ReadyTimeout); err != nil {
			return resp, err
		}

		var stdout, stderr bytes.Buffer
		args := serveStopArgs(req)
		cliErr := browseragent.HandleCLI(args, req.Env, &stdout, &stderr)
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		if cliErr != nil {
			resp.CLIErr = cliErr.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		resp.ExitCode = 0

		downWait := req.ShutdownWait
		if downWait <= 0 {
			downWait = 8 * time.Second
		}
		if err := waitHealthDown(srv.BaseURL, downWait); err != nil {
			return resp, fmt.Errorf("daemon still healthy after serve --stop: %w", err)
		}
		resp.HealthDown = true

		select {
		case <-srv.done:
			resp.DaemonExited = true
		case <-time.After(downWait):
		}

		_, err = os.Stat(daemonMetaPath(req.BaseDir))
		if err == nil {
			resp.DaemonMetaExists = true
			return resp, nil
		}
		if os.IsNotExist(err) {
			resp.DaemonMetaExists = false
		} else {
			return resp, err
		}

		low := strings.ToLower(resp.Stderr)
		resp.StopMessageSeen = strings.Contains(low, "stop") ||
			strings.Contains(low, "stopped") ||
			strings.Contains(low, "shutdown") ||
			strings.Contains(low, "kill")
		resp.HasANSI = containsANSI(resp.Stderr)
		return resp, nil

	case StopOpNotRunning:
		var stdout, stderr bytes.Buffer
		args := serveStopArgs(req)
		cliErr := browseragent.HandleCLI(args, req.Env, &stdout, &stderr)
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		if cliErr != nil {
			resp.CLIErr = cliErr.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		resp.ExitCode = 0

		low := strings.ToLower(resp.Stderr)
		resp.WarningSeen = strings.Contains(low, "warning:") &&
			strings.Contains(low, "no daemon running")
		resp.HasANSI = containsANSI(resp.Stderr)
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown StopOp %q", req.StopOp)
	}
}

func runModeConflictMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ModeConflictOp == "" {
		t.Fatal("ModeConflictOp must be set by leaf Setup")
	}

	var stdout, stderr bytes.Buffer
	args := req.CLIArgs
	if len(args) == 0 {
		switch req.ModeConflictOp {
		case ModeConflictOpStopStatus:
			args = []string{
				"serve",
				"--stop",
				"--status",
				"--base-dir", req.BaseDir,
			}
		case ModeConflictOpStopKillExisting:
			args = []string{
				"serve",
				"--stop",
				"--kill-existing",
				"--no-open-chrome",
				"--no-agent-run",
				"--base-dir", req.BaseDir,
			}
		default:
			return nil, fmt.Errorf("unknown ModeConflictOp %q", req.ModeConflictOp)
		}
	}

	cliErr := browseragent.HandleCLI(args, req.Env, &stdout, &stderr)
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	} else {
		resp.ExitCode = 0
	}

	combined := strings.ToLower(resp.Stderr + "\n" + resp.CLIErr)
	resp.MutualExclSeen = strings.Contains(combined, "mutually exclusive") ||
		(strings.Contains(combined, "--stop") &&
			strings.Contains(combined, "--status") &&
			strings.Contains(combined, "--kill-existing"))
	resp.HasANSI = containsANSI(resp.Stderr)
	return resp, nil
}

func runUnknownFlagMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	args := req.CLIArgs
	if len(args) == 0 {
		args = []string{"serve", "--foo", "--base-dir", req.BaseDir}
	}

	cliErr := browseragent.HandleCLI(args, req.Env, &stdout, &stderr)
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	} else {
		resp.ExitCode = 0
	}

	combined := strings.ToLower(resp.Stderr + "\n" + resp.CLIErr)
	resp.UnknownFlagSeen = strings.Contains(combined, "unrecognized flag") ||
		strings.Contains(combined, "unknown flag")
	return resp, nil
}

func runHelpMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	args := req.CLIArgs
	if len(args) == 0 {
		args = []string{"serve", "--help"}
	}

	cliErr := browseragent.HandleCLI(args, req.Env, &stdout, &stderr)
	resp := &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		HelpText: stdout.String(),
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
		return resp, nil
	}
	resp.ExitCode = 0
	return resp, nil
}

func runColorMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ColorOp == "" {
		t.Fatal("ColorOp must be set by leaf Setup")
	}

	var stdout, stderr bytes.Buffer
	args := req.CLIArgs
	if len(args) == 0 {
		switch req.ColorOp {
		case ColorOpForceOn:
			args = []string{"serve", "--color", "--stop", "--base-dir", req.BaseDir}
		case ColorOpNoColorFlag:
			args = []string{"serve", "--no-color", "--stop", "--base-dir", req.BaseDir}
		case ColorOpNoColorEnv:
			args = []string{"serve", "--stop", "--base-dir", req.BaseDir}
		case ColorOpColorConflict:
			args = []string{"serve", "--color", "--no-color", "--stop", "--base-dir", req.BaseDir}
		default:
			return nil, fmt.Errorf("unknown ColorOp %q", req.ColorOp)
		}
	}

	cliErr := browseragent.HandleCLI(args, req.Env, &stdout, &stderr)
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	} else {
		resp.ExitCode = 0
	}
	resp.HasANSI = containsANSI(resp.Stderr)

	combined := strings.ToLower(resp.Stderr + "\n" + resp.CLIErr)
	resp.ColorConflictSeen = strings.Contains(combined, "cannot be specified together") ||
		strings.Contains(combined, "mutually exclusive") ||
		(strings.Contains(combined, "--color") && strings.Contains(combined, "--no-color"))
	resp.WarningSeen = strings.Contains(strings.ToLower(resp.Stderr), "warning:")
	return resp, nil
}

func runRegressionMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
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
		done <- browseragent.HandleCLI(args, req.Env, &stdout, &stderr)
	}()

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
		HealthDown: true,
	}
	low := strings.ToLower(resp.Stderr)
	resp.KillMessageSeen = strings.Contains(low, "kill") ||
		strings.Contains(low, "shutdown") ||
		strings.Contains(low, "existing") ||
		strings.Contains(low, "stopped")

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
			resp.CLIErr = err.Error()
			resp.ExitCode = 1
			return resp, err
		}
		resp.ExitCode = 0
		return resp, nil
	case <-time.After(downWait):
		return resp, fmt.Errorf("HandleCLI did not return after shutting down second serve")
	}
}

func serveStopArgs(req *Request) []string {
	if len(req.CLIArgs) > 0 {
		return req.CLIArgs
	}
	args := []string{
		"serve",
		"--stop",
		"--no-open-chrome",
		"--no-agent-run",
		"--base-dir", req.BaseDir,
	}
	if req.Addr != "" {
		args = append(args, "--addr", req.Addr)
	}
	return args
}

// --- RunDaemon harness (phase 6 parity) ---

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

func daemonMetaPath(baseDir string) string {
	return filepath.Join(baseDir, "server.json")
}

func containsANSI(s string) bool {
	return strings.Contains(s, "\x1b")
}

var (
	_ = sync.Mutex{}
	_ = io.Discard
)
```