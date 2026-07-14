# browser-agent session side-commands — auto-resolve --addr from server.json

Bug fix from LOOP `docs/loops/LOOP_2026-07-14_session-info-addr-mismatch.md`
(SYMPTOM CONFIRMED).

**Root cause:** `serve --status` reads `{base-dir}/server.json` for the actual daemon
addr, but `session info|eval|run|logs|screenshot|cdp` default to
`http://127.0.0.1:43761` when `--addr` is omitted → **404 session not found** when
the daemon listens on an ephemeral port recorded in meta.

**Fix:** When `--addr` is not set on session side-commands, resolve the control base
URL from `server.json` under `--base-dir` (default `~/.tmp/browser-agent`), matching
`QueryDaemonStatus` / `EnsureDaemon` / `tryReuseDaemon`. Explicit `--addr` still
wins.

**No real Chrome.** **No real agent-run.** In-process `RunDaemon` on ephemeral
`127.0.0.1:0` with temp `BaseDir` and `server.json`.

| Surface | What is under test |
|---------|-------------------|
| `session info` | Omit `--addr`; pass `--base-dir` with meta → hits correct daemon |
| `session info --addr` | Explicit addr overrides meta (regression) |
| `session eval` | Omit `--addr`; job POST must not 404 unknown session |
| Fallback | No `server.json` → default `43761` (connection error OK) |

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Host** (`RunDaemon`) binds the control HTTP server on an ephemeral port,
writes `{BaseDir}/server.json` (`DaemonMeta` with `addr`, `base_url`, `pid`,
`base_dir`), and serves `POST /v1/sessions` for session creation.

**Operator CLI** — nested session side-commands (`session info`, `session eval`, …)
accept `--session-id`, optional `--addr`, and (after fix) `--base-dir`. When
`--addr` is omitted, the CLI should read meta from `{baseDir}/server.json` and use
`base_url` or `http://{addr}` when the recorded pid is alive and health is OK.

**Status probe** (`QueryDaemonStatus` / `serve --status`) already resolves addr from
meta — this tree proves side-commands follow the same discovery path.

**Test Client** starts `RunDaemon`, creates a session via `POST /v1/sessions`,
invokes `HandleCLI` with controlled argv (with/without `--addr`, `--base-dir`), and
asserts stdout / error text.

```text
RunDaemon(:0, BaseDir) -> server.json (addr != 43761)
POST /v1/sessions -> sess-xxxxxx

# BUG (RED)
HandleCLI session info --session-id X --base-dir BaseDir   # no --addr
  -> GET http://127.0.0.1:43761/v1/session -> 404 session not found

# FIX (GREEN)
HandleCLI session info --session-id X --base-dir BaseDir
  -> read server.json -> GET http://127.0.0.1:<ephemeral>/v1/session -> exit 0

HandleCLI session eval --session-id X --base-dir BaseDir '1+1'
  -> POST correct /v1/jobs (not 404 unknown session)

HandleCLI session info --session-id X --addr <meta> --base-dir BaseDir
  -> explicit --addr wins (regression)
```

## Decision Tree

```
browser-agent-session-addr-resolve
├── info/                                    [session info addr resolution]
│   ├── from-server-json/                      ephemeral port; no --addr; exit 0 + session_id
│   └── explicit-addr-overrides/               --addr matches daemon; exit 0 (regression)
├── eval/                                    [same resolution for eval]
│   └── from-server-json/                      no --addr; must NOT 404 unknown session
└── fallback/                                [no server.json]
    └── default-addr-when-no-meta/             no meta + no --addr → default 43761 (conn err OK)
```

### Parameter significance (high → low)

1. **Side-command** — `info` vs `eval` (different HTTP paths).
2. **Addr source** — meta (`server.json`) vs explicit `--addr` vs default fallback.
3. **Daemon setup** — ephemeral `RunDaemon` vs no meta (fallback only).

## Test Index

| Leaf | Scenario |
|------|----------|
| `info/from-server-json` | `RunDaemon` `:0`; create session; `session info --session-id X --base-dir <dir>` **no --addr** → exit 0; stdout contains `session_id` |
| `info/explicit-addr-overrides` | Same daemon; `--addr` matching meta → exit 0; stdout contains session id |
| `eval/from-server-json` | Same; `session eval '1+1'` without `--addr` → must **not** contain `session not found` / `unknown session` |
| `fallback/default-addr-when-no-meta` | No `server.json`; `session info` without `--addr` → connection error OK; must **not** imply meta-based resolution |

**Leaf count: 4**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-addr-resolve
doctest test ./tests/browser-agent-session-addr-resolve   # RED until implementer lands fix
# regressions after fix:
doctest test ./tests/browser-agent-product-hardening/session-info-cli/...
doctest test ./tests/browser-agent-cli-react/cli-sidecmd/...
```

### Implementer contract (authoritative for GREEN)

```text
// Resolution order for session side-commands when --addr omitted:
// 1. --addr flag if set -> normalizeAddr(flag)
// 2. Else read {baseDir}/server.json -> base_url or http://{addr} when pid alive + health OK
// 3. Else default http://127.0.0.1:43761

// Add --base-dir to session side-command flags (default ~/.tmp/browser-agent) if missing.
// Package helper e.g. ResolveControlBaseURL(baseDir, addrFlag string) string

func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Sidecmd — nested session command under test.
const (
	SidecmdInfo = "info"
	SidecmdEval = "eval"
)

// AddrSource — how the CLI should resolve the control base URL.
const (
	AddrFromServerJSON = "from-server-json"
	AddrExplicit       = "explicit-addr"
	AddrDefaultNoMeta  = "default-no-meta"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Sidecmd    string // info | eval
	AddrSource string // from-server-json | explicit-addr | default-no-meta

	ModuleRoot string
	BaseDir    string

	// Live daemon (from-server-json + explicit-addr leaves).
	StartDaemon bool
	Addr        string // bound listen addr after RunDaemon
	BaseURL     string

	SessionID string
	EvalExpr  string

	// CLI argv control (Run builds argv from these).
	OmitAddr     bool // true = do not pass --addr (bug repro path)
	PassBaseDir  bool // pass --base-dir <BaseDir>
	ExplicitAddr string // when AddrSource=explicit-addr

	ReadyTimeout    time.Duration
	MaxDispatchWait time.Duration
	CLIEnv          map[string]string
}

// Response holds daemon + CLI outcomes.
type Response struct {
	BaseURL string
	Addr    string

	DaemonMetaPath   string
	DaemonMetaExists bool
	DaemonMeta       browseragent.DaemonMeta

	SessionID string

	Stdout string
	Stderr string
	ExitCode int
	CLIErr   string
	ErrText  string

	DispatchTimedOut bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Sidecmd == "" {
		t.Fatal("Sidecmd must be set by grouping/leaf Setup")
	}
	if req.AddrSource == "" {
		t.Fatal("AddrSource must be set by leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	if req.EvalExpr == "" {
		req.EvalExpr = "1+1"
	}
	if req.MaxDispatchWait <= 0 {
		req.MaxDispatchWait = 12 * time.Second
	}

	resp := &Response{
		DaemonMetaPath: daemonMetaPath(req.BaseDir),
	}

	switch req.AddrSource {
	case AddrFromServerJSON, AddrExplicit:
		return runWithDaemon(t, req, resp)
	case AddrDefaultNoMeta:
		return runFallbackNoMeta(t, req, resp)
	default:
		return nil, fmt.Errorf("unknown AddrSource %q", req.AddrSource)
	}
}

func runWithDaemon(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return resp, err
	}
	defer cleanup()

	resp.BaseURL = srv.BaseURL
	resp.Addr = srv.Addr
	req.BaseURL = srv.BaseURL
	req.Addr = srv.Addr

	meta, _, exists, readErr := readDaemonMetaFile(req.BaseDir)
	resp.DaemonMetaExists = exists
	resp.DaemonMeta = meta
	if readErr != nil && !os.IsNotExist(readErr) {
		return resp, readErr
	}
	if !exists {
		return resp, fmt.Errorf("server.json missing under %s after RunDaemon", req.BaseDir)
	}
	if meta.Addr == "127.0.0.1:43761" || strings.HasSuffix(meta.Addr, ":43761") {
		return resp, fmt.Errorf("ephemeral daemon must not bind default :43761; got %q", meta.Addr)
	}

	sid, err := createSessionHTTP(srv.BaseURL, "")
	if err != nil {
		return resp, err
	}
	resp.SessionID = sid
	req.SessionID = sid

	args := buildSidecmdArgs(req)
	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)
	return resp, err
}

func runFallbackNoMeta(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	metaPath := daemonMetaPath(req.BaseDir)
	_, statErr := os.Stat(metaPath)
	if statErr == nil {
		return resp, fmt.Errorf("fallback leaf requires absent server.json; found %s", metaPath)
	}
	if !os.IsNotExist(statErr) {
		return resp, statErr
	}
	resp.DaemonMetaExists = false

	req.SessionID = "sess-fallback-no-meta"
	args := buildSidecmdArgs(req)
	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)
	return resp, err
}

func buildSidecmdArgs(req *Request) []string {
	args := []string{"session", req.Sidecmd}
	if req.PassBaseDir && req.BaseDir != "" {
		args = append(args, "--base-dir", req.BaseDir)
	}
	if req.SessionID != "" {
		args = append(args, "--session-id", req.SessionID)
	}
	switch req.AddrSource {
	case AddrExplicit:
		addr := req.ExplicitAddr
		if addr == "" {
			addr = req.BaseURL
		}
		if addr == "" {
			addr = "http://" + req.Addr
		}
		host, portStr, _ := net.SplitHostPort(strings.TrimPrefix(strings.TrimPrefix(addr, "https://"), "http://"))
		args = append(args, "--host", host, "--server-port", portStr)
	case AddrFromServerJSON:
		// Intentionally omit --addr (bug repro / fix verify path).
	case AddrDefaultNoMeta:
		// Omit --addr; no server.json in BaseDir.
	}
	if req.Sidecmd == SidecmdInfo {
		args = append(args, "--json")
	}
	if req.Sidecmd == SidecmdEval {
		args = append(args, req.EvalExpr)
	}
	return args
}

func mergeCLIResponse(resp *Response, cli *Response) {
	if cli == nil {
		return
	}
	resp.Stdout = cli.Stdout
	resp.Stderr = cli.Stderr
	resp.ExitCode = cli.ExitCode
	resp.CLIErr = cli.CLIErr
	resp.ErrText = cli.ErrText
	resp.DispatchTimedOut = cli.DispatchTimedOut
}

func invokeHandleCLI(t *testing.T, req *Request, args []string) (*Response, error) {
	t.Helper()
	maxWait := req.MaxDispatchWait
	var stdout, stderr bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- browseragent.HandleCLI(args, req.CLIEnv, &stdout, &stderr)
	}()

	resp := &Response{}
	select {
	case err := <-done:
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		if err != nil {
			resp.CLIErr = err.Error()
			resp.ErrText = err.Error()
			resp.ExitCode = 1
		} else {
			resp.ExitCode = 0
		}
		return resp, nil
	case <-time.After(maxWait):
		resp.DispatchTimedOut = true
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		return resp, fmt.Errorf("HandleCLI timed out after %v: args=%v", maxWait, args)
	}
}

// --- RunDaemon harness (reuse phase6 / phase7 patterns) ---

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

func createSessionHTTP(baseURL, sessionID string) (string, error) {
	body := map[string]string{}
	if sessionID != "" {
		body["session_id"] = sessionID
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/sessions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("POST /v1/sessions status=%d body=%s", res.StatusCode, strings.TrimSpace(string(out)))
	}
	var parsed map[string]string
	if err := json.Unmarshal(out, &parsed); err != nil {
		return "", fmt.Errorf("parse POST /v1/sessions: %w", err)
	}
	sid := parsed["session_id"]
	if sid == "" {
		return "", fmt.Errorf("POST /v1/sessions missing session_id")
	}
	return sid, nil
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

```