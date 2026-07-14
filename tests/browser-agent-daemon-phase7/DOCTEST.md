# browser-agent daemon Phase 7 — serve --status (read-only)

Phase 7 adds **`browser-agent serve --status`** — a read-only operator probe.
No spawn, no Chrome, no writes to `server.json`.

Package API:

- **`QueryDaemonStatus(baseDir)`** → `DaemonStatus` (running flag, pid, addr,
  base_url, base_dir, uptime, started_at, sessions from `GET /v1/sessions` when
  running)
- **`FormatDaemonStatus(w, st)`** → pretty table on stdout

CLI: `cliServe` with `--status` calls `QueryDaemonStatus`, `FormatDaemonStatus`
to stdout, and exits **0** even when the daemon is not running.

**No real Chrome.** **No real agent-run.** In-process `RunDaemon` on ephemeral
`127.0.0.1:0` with temp `BaseDir` for running leaves; stale/not-running leaves
use filesystem fixtures only.

| Surface | What is under test |
|---------|-------------------|
| `QueryDaemonStatus` | Running / not-running / stale-pid detection |
| `FormatDaemonStatus` | Pretty table markers (`status`, `sessions`, headers) |
| `serve --status` | Read-only CLI; exit 0; stdout table; no daemon spawn |

Depends on Phase 5 (`RunDaemon`, `DaemonMeta`, `server.json`) and Phase 3
(`POST /v1/sessions`, `GET /v1/sessions`).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Status Query** (`QueryDaemonStatus`) reads `{BaseDir}/server.json` when
present, checks whether the recorded **pid** is alive and the control plane
responds, and when running fetches **`GET /v1/sessions`** for session snapshots.
Missing meta → `Running=false`. Stale meta (dead pid) → `Running=false` without
mutating `server.json`.

**Status Formatter** (`FormatDaemonStatus`) renders operator-facing table output
to an `io.Writer` — includes daemon fields and a sessions section when running.

**Operator CLI** — `browser-agent serve --status` queries and prints status, then
**returns immediately** (does not call `RunDaemon`). Exit code is **0** for both
running and not-running.

**Test Client** starts `RunDaemon` only for running scenarios, creates sessions
via `POST /v1/sessions`, snapshots `server.json` before/after status probes, and
asserts no meta mutation and no extra processes.

```text
QueryDaemonStatus(baseDir)
  no server.json -> Running=false
  stale pid      -> Running=false (meta unchanged)
  live daemon    -> Running=true + sessions from GET /v1/sessions

FormatDaemonStatus(stdout, st) -> table with status/sessions markers

HandleCLI serve --status -> QueryDaemonStatus + FormatDaemonStatus -> exit 0
```

## Decision Tree

```
browser-agent-daemon-phase7
├── query-status/                        [QueryDaemonStatus helper]
│   ├── running-with-sessions/             live daemon + sessions populated
│   ├── not-running-no-file/               no server.json -> Running=false
│   └── stale-pid/                         dead pid in meta -> Running=false
├── format-status/                       [FormatDaemonStatus helper]
│   └── pretty-output-markers/             stdout contains status/sessions/headers
└── cli-status/                          [HandleCLI serve --status]
    ├── running-exit-0/                    running daemon; stdout table; exit 0
    └── not-running-exit-0/                no daemon; stdout not-running; exit 0
```

### Parameter significance (high → low)

1. **Entry surface** — `QueryDaemonStatus` vs `FormatDaemonStatus` vs CLI.
2. **Daemon state** — running with sessions vs absent meta vs stale pid.
3. **Side effects** — meta bytes unchanged; CLI must not spawn daemon.

## Test Index

| Leaf | Scenario |
|------|----------|
| `query-status/running-with-sessions` | Live daemon; `Running=true`; sessions match HTTP list |
| `query-status/not-running-no-file` | No `server.json`; `Running=false`; no file created |
| `query-status/stale-pid` | Stale meta with dead pid; `Running=false`; meta unchanged |
| `format-status/pretty-output-markers` | `FormatDaemonStatus` table markers |
| `cli-status/running-exit-0` | `serve --status` with live daemon; exit 0; daemon stays up |
| `cli-status/not-running-exit-0` | `serve --status` without daemon; exit 0; no meta write |

**Leaf count: 6**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase7
doctest test ./tests/browser-agent-daemon-phase7
# After implementer lands phase 7:
doctest test ./tests/browser-agent-daemon-phase6
doctest test ./tests/browser-agent-daemon-phase5
doctest test ./tests/browser-agent/...
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until
implementer lands phase 7):

- `DaemonStatus` struct (`Running`, `PID`, `Addr`, `BaseURL`, `BaseDir`,
  `Uptime`, `StartedAt`, `Sessions`)
- `QueryDaemonStatus(baseDir string) (DaemonStatus, error)`
- `FormatDaemonStatus(w io.Writer, st DaemonStatus) error`
- `HandleCLI serve --status` — read-only; exit 0 when not running

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
	ModeQueryStatus  = "query-status"
	ModeFormatStatus = "format-status"
	ModeCLIStatus    = "cli-status"
)

// QueryStatusOp — QueryDaemonStatus probes.
const (
	QueryStatusOpRunningWithSessions = "running-with-sessions"
	QueryStatusOpNotRunningNoFile    = "not-running-no-file"
	QueryStatusOpStalePID            = "stale-pid"
)

// FormatStatusOp — FormatDaemonStatus probes.
const (
	FormatStatusOpPrettyMarkers = "pretty-output-markers"
)

// CLIStatusOp — serve --status probes.
const (
	CLIStatusOpRunningExit0    = "running-exit-0"
	CLIStatusOpNotRunningExit0 = "not-running-exit-0"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	QueryStatusOp  string
	FormatStatusOp string
	CLIStatusOp    string

	// Session created on running leaves.
	SessionID string

	// Stale meta fixture pid (dead).
	StalePID int

	ReadyTimeout time.Duration
}

// Response holds status query / format / CLI outcomes.
type Response struct {
	Status      browseragent.DaemonStatus
	QueryErr    string
	Formatted   string
	FormatErr   string

	BaseURL string
	Addr    string

	SessionIDsFromHTTP []string
	StatusSessionIDs   []string

	DaemonMetaPath      string
	DaemonMetaBefore    []byte
	DaemonMetaAfter     []byte
	DaemonMetaBeforeOK  bool
	DaemonMetaAfterOK   bool
	DaemonMetaBeforeHit bool
	DaemonMetaAfterHit  bool

	DaemonHealthyAfter bool

	Stdout string
	Stderr string
	CLIErr string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeQueryStatus:
		return runQueryStatusMode(t, req)
	case ModeFormatStatus:
		return runFormatStatusMode(t, req)
	case ModeCLIStatus:
		return runCLIStatusMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runQueryStatusMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.QueryStatusOp == "" {
		t.Fatal("QueryStatusOp must be set by leaf Setup")
	}

	resp := &Response{
		DaemonMetaPath: daemonMetaPath(req.BaseDir),
	}

	switch req.QueryStatusOp {
	case QueryStatusOpRunningWithSessions:
		srv, cleanup, err := startDaemonServer(t, req)
		if err != nil {
			return resp, err
		}
		defer cleanup()

		resp.BaseURL = srv.BaseURL
		resp.Addr = srv.Addr

		if err := createSessionHTTP(srv.BaseURL, req.SessionID); err != nil {
			return resp, err
		}
		ids, err := listSessionsHTTP(srv.BaseURL)
		if err != nil {
			return resp, err
		}
		resp.SessionIDsFromHTTP = ids

		before, ok, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaBefore = before
		resp.DaemonMetaBeforeOK = ok
		resp.DaemonMetaBeforeHit = ok

		st, qerr := browseragent.QueryDaemonStatus(req.BaseDir)
		if qerr != nil {
			resp.QueryErr = qerr.Error()
			return resp, qerr
		}
		resp.Status = st
		resp.StatusSessionIDs = sessionIDsFromStatus(st)

		after, ok2, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaAfter = after
		resp.DaemonMetaAfterOK = ok2
		resp.DaemonMetaAfterHit = ok2

		resp.DaemonHealthyAfter = healthOK(srv.BaseURL)
		return resp, nil

	case QueryStatusOpNotRunningNoFile:
		before, ok, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaBefore = before
		resp.DaemonMetaBeforeOK = ok
		resp.DaemonMetaBeforeHit = ok

		st, qerr := browseragent.QueryDaemonStatus(req.BaseDir)
		if qerr != nil {
			resp.QueryErr = qerr.Error()
			return resp, qerr
		}
		resp.Status = st

		after, ok2, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaAfter = after
		resp.DaemonMetaAfterOK = ok2
		resp.DaemonMetaAfterHit = ok2
		return resp, nil

	case QueryStatusOpStalePID:
		stalePID := req.StalePID
		if stalePID == 0 {
			stalePID = 999999999
		}
		meta := browseragent.DaemonMeta{
			PID:       stalePID,
			Addr:      "127.0.0.1:59999",
			BaseURL:   "http://127.0.0.1:59999",
			BaseDir:   req.BaseDir,
			StartedAt: time.Now().Add(-2 * time.Hour),
		}
		metaPath := daemonMetaPath(req.BaseDir)
		if err := browseragent.WriteDaemonMeta(metaPath, meta); err != nil {
			return resp, err
		}

		before, ok, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaBefore = before
		resp.DaemonMetaBeforeOK = ok
		resp.DaemonMetaBeforeHit = ok

		st, qerr := browseragent.QueryDaemonStatus(req.BaseDir)
		if qerr != nil {
			resp.QueryErr = qerr.Error()
			return resp, qerr
		}
		resp.Status = st

		after, ok2, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaAfter = after
		resp.DaemonMetaAfterOK = ok2
		resp.DaemonMetaAfterHit = ok2
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown QueryStatusOp %q", req.QueryStatusOp)
	}
}

func runFormatStatusMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.FormatStatusOp == "" {
		t.Fatal("FormatStatusOp must be set by leaf Setup")
	}

	resp := &Response{}

	switch req.FormatStatusOp {
	case FormatStatusOpPrettyMarkers:
		srv, cleanup, err := startDaemonServer(t, req)
		if err != nil {
			return resp, err
		}
		defer cleanup()

		resp.BaseURL = srv.BaseURL
		resp.Addr = srv.Addr

		if err := createSessionHTTP(srv.BaseURL, req.SessionID); err != nil {
			return resp, err
		}

		st, qerr := browseragent.QueryDaemonStatus(req.BaseDir)
		if qerr != nil {
			resp.QueryErr = qerr.Error()
			return resp, qerr
		}
		resp.Status = st

		var buf bytes.Buffer
		if err := browseragent.FormatDaemonStatus(&buf, st); err != nil {
			resp.FormatErr = err.Error()
			return resp, err
		}
		resp.Formatted = buf.String()
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown FormatStatusOp %q", req.FormatStatusOp)
	}
}

func runCLIStatusMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLIStatusOp == "" {
		t.Fatal("CLIStatusOp must be set by leaf Setup")
	}

	resp := &Response{
		DaemonMetaPath: daemonMetaPath(req.BaseDir),
	}

	switch req.CLIStatusOp {
	case CLIStatusOpRunningExit0:
		srv, cleanup, err := startDaemonServer(t, req)
		if err != nil {
			return resp, err
		}
		defer cleanup()

		resp.BaseURL = srv.BaseURL
		resp.Addr = srv.Addr

		if err := createSessionHTTP(srv.BaseURL, req.SessionID); err != nil {
			return resp, err
		}

		before, ok, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaBefore = before
		resp.DaemonMetaBeforeOK = ok
		resp.DaemonMetaBeforeHit = ok

		var stdout, stderr bytes.Buffer
		args := []string{
			"serve",
			"--status",
			"--base-dir", req.BaseDir,
		}
		cliErr := browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)
		if cliErr != nil {
			resp.CLIErr = cliErr.Error()
		}
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()

		after, ok2, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaAfter = after
		resp.DaemonMetaAfterOK = ok2
		resp.DaemonMetaAfterHit = ok2

		resp.DaemonHealthyAfter = healthOK(srv.BaseURL)
		return resp, cliErr

	case CLIStatusOpNotRunningExit0:
		before, ok, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaBefore = before
		resp.DaemonMetaBeforeOK = ok
		resp.DaemonMetaBeforeHit = ok

		var stdout, stderr bytes.Buffer
		args := []string{
			"serve",
			"--status",
			"--base-dir", req.BaseDir,
		}
		cliErr := browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)
		if cliErr != nil {
			resp.CLIErr = cliErr.Error()
		}
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()

		after, ok2, err := readMetaBytes(req.BaseDir)
		if err != nil {
			return resp, err
		}
		resp.DaemonMetaAfter = after
		resp.DaemonMetaAfterOK = ok2
		resp.DaemonMetaAfterHit = ok2
		return resp, cliErr

	default:
		return nil, fmt.Errorf("unknown CLIStatusOp %q", req.CLIStatusOp)
	}
}

// --- harness helpers ---

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

func createSessionHTTP(baseURL, sessionID string) error {
	body, err := json.Marshal(map[string]string{"session_id": sessionID})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/sessions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("POST /v1/sessions status=%d body=%s", res.StatusCode, strings.TrimSpace(string(out)))
	}
	return nil
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

func sessionIDsFromStatus(st browseragent.DaemonStatus) []string {
	ids := make([]string, 0, len(st.Sessions))
	for _, snap := range st.Sessions {
		ids = append(ids, snap.SessionID)
	}
	return ids
}

func daemonMetaPath(baseDir string) string {
	return filepath.Join(baseDir, "server.json")
}

func readMetaBytes(baseDir string) ([]byte, bool, error) {
	path := daemonMetaPath(baseDir)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return raw, true, nil
}

var (
	_ = sync.Mutex{}
	_ = io.Discard
)
```