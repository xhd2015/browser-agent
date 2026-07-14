# browser-agent serve --status rich output (version + extension + connected)

Enriches **`browser-agent serve --status`** with daemon version, embedded extension
bundle metadata, and a **Connected** column in the sessions table.

Package API extensions on top of Phase 7:

- **`DaemonStatus`** adds `DaemonVersion`, `ExtensionVersion`, `ExtensionMD5`,
  `ExtensionPath`
- **`QueryDaemonStatus`** populates version/extension fields (running and not-running)
- **`FormatDaemonStatus`** prints version block, extension block, and Connected column

CLI: `serve --status` remains read-only; exit **0** when not running.

**No real Chrome.** **No real agent-run.** Ephemeral `127.0.0.1:0` listen and
isolated `HOME` for canonical extension paths.

| Surface | What is under test |
|---------|-------------------|
| `QueryDaemonStatus` | Rich fields when running / not-running |
| `FormatDaemonStatus` | Version block, extension block, Connected column |
| `serve --status` | Integration stdout; phase7 parity + new markers |

Depends on Phase 7 status APIs and `EnsureCanonicalExtension` / `ReadBundleSumFromDir`.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Status Query** (`QueryDaemonStatus`) reads `{BaseDir}/server.json`, probes
the control plane when the pid is alive, and fetches **`GET /v1/sessions`**. It also
resolves **daemon version** (meta → health → effective) and **embedded extension**
metadata (session bundle or canonical extract via `EnsureCanonicalExtension`).

**Status Formatter** (`FormatDaemonStatus`) renders operator-facing table output:
uptime, **Version**, **Extension (embedded)** block, and sessions with **Connected**
(`yes`/`no`).

**Operator CLI** — `browser-agent serve --status` queries and prints rich status,
then returns immediately (does not call `RunDaemon`). Exit **0** for running and
not-running.

**Test Client** starts `RunDaemon` on ephemeral port for running scenarios, sets
isolated `HOME` for canonical extension layout, snapshots `server.json` before/after
probes, and asserts read-only behavior.

```text
QueryDaemonStatus(baseDir)
  running     -> DaemonVersion + extension fields + sessions
  not running -> extension fields from canonical extract (no daemon version)

FormatDaemonStatus(stdout, st) -> Version + Extension block + Connected column

HandleCLI serve --status -> QueryDaemonStatus + FormatDaemonStatus -> exit 0
```

## Decision Tree

```
browser-agent-serve-status-rich
├── query-status/                        [QueryDaemonStatus rich fields]
│   ├── running-populates-version/         DaemonVersion non-empty when daemon up
│   ├── running-populates-extension/       ExtensionPath under extensions/browser-agent/
│   └── not-running-populates-extension/   ExtensionPath when meta absent
├── format-status/                       [FormatDaemonStatus stdout]
│   ├── running-shows-version-block/       Version + Extension (embedded) markers
│   ├── running-shows-connected-column/    Connected header + yes/no
│   ├── running-session-row/               session id + phase in table
│   └── not-running-shows-extension/       extension block without running status
└── cli-status/                          [HandleCLI serve --status]
    ├── running-exit-0/                    integration; rich markers + phase7 parity
    └── not-running-exit-0/                extension block when not running
```

### Parameter significance (high → low)

1. **Entry surface** — `QueryDaemonStatus` vs `FormatDaemonStatus` vs CLI.
2. **Daemon state** — running with session vs not-running (no meta).
3. **Rich fields** — version, extension bundle, Connected column.

## Test Index

| Leaf | Scenario |
|------|----------|
| `query-status/running-populates-version` | `DaemonVersion` non-empty when daemon healthy |
| `query-status/running-populates-extension` | `ExtensionPath` contains `extensions/browser-agent/` |
| `query-status/not-running-populates-extension` | `ExtensionPath` set without `server.json` |
| `format-status/running-shows-version-block` | `Version:`, `Extension (embedded)`, `md5`, `path` |
| `format-status/running-shows-connected-column` | `Connected` header; `yes` or `no` for fixture session |
| `format-status/running-session-row` | Created session id and phase appear in table |
| `format-status/not-running-shows-extension` | Extension block without `Status:   running` |
| `cli-status/running-exit-0` | `serve --status` rich stdout; exit 0; meta unchanged |
| `cli-status/not-running-exit-0` | `serve --status` extension block; exit 0; no meta write |

**Leaf count: 9**

## How to Run

```sh
doctest vet ./tests/browser-agent-serve-status-rich
doctest test ./tests/browser-agent-serve-status-rich
# After implementer lands rich status:
doctest test ./tests/browser-agent-daemon-phase7
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (**RED** until
implementer extends status):

- `DaemonStatus` adds `DaemonVersion`, `ExtensionVersion`, `ExtensionMD5`, `ExtensionPath`
- `QueryDaemonStatus` populates rich fields per REQUIREMENT-DESIGN-serve-status-rich.md
- `FormatDaemonStatus` prints version/extension blocks and Connected column
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

const (
	ModeQueryStatus  = "query-status"
	ModeFormatStatus = "format-status"
	ModeCLIStatus    = "cli-status"
)

const (
	QueryStatusOpRunningPopulatesVersion    = "running-populates-version"
	QueryStatusOpRunningPopulatesExtension  = "running-populates-extension"
	QueryStatusOpNotRunningPopulatesExt     = "not-running-populates-extension"
)

const (
	FormatStatusOpRunningShowsVersionBlock    = "running-shows-version-block"
	FormatStatusOpRunningShowsConnectedColumn = "running-shows-connected-column"
	FormatStatusOpRunningSessionRow           = "running-session-row"
	FormatStatusOpNotRunningShowsExtension      = "not-running-shows-extension"
)

const (
	CLIStatusOpRunningExit0    = "running-exit-0"
	CLIStatusOpNotRunningExit0 = "not-running-exit-0"
)

type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	TestHome   string
	Addr       string

	QueryStatusOp  string
	FormatStatusOp string
	CLIStatusOp    string

	SessionID string
	DaemonVersion string

	ReadyTimeout time.Duration
}

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
	case QueryStatusOpRunningPopulatesVersion, QueryStatusOpRunningPopulatesExtension:
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

	case QueryStatusOpNotRunningPopulatesExt:
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
	case FormatStatusOpRunningShowsVersionBlock,
		FormatStatusOpRunningShowsConnectedColumn,
		FormatStatusOpRunningSessionRow:
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

	case FormatStatusOpNotRunningShowsExtension:
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
	daemonVer := strings.TrimSpace(req.DaemonVersion)
	if daemonVer == "" {
		daemonVer = browseragent.ClientVersion()
	}
	cfg := browseragent.DaemonConfig{
		Addr:          addr,
		BaseDir:       req.BaseDir,
		Stdout:        io.Discard,
		Stderr:        io.Discard,
		DaemonVersion: daemonVer,
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