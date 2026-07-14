# browser-agent session list — CLI listing via GET /v1/sessions

Greenfield feature: **`browser-agent session list`** fetches live session snapshots from
`GET {baseURL}/v1/sessions` via `ResolveControlBaseURL` (same addr discovery as other
session side-commands). Human table on stdout (default); `--json` emits raw JSON array.

**Daemon unreachable → exit 0** with `warning: daemon not running in <base-dir>` on stderr
and empty human output (`0 sessions`) or `[]` with `--json`. **No disk-only orphan dirs**
in v1 (registry/live API only).

**No real Chrome.** **No real agent-run.** In-process `RunDaemon` on ephemeral
`127.0.0.1:0` with temp `BaseDir`; fake extension WS hello for connected column leaf.

| Surface | What is under test |
|---------|-------------------|
| `session list` CLI | Human table: Session ID, Phase, Connected; trailing `N sessions` |
| Addr resolution | Omit `--addr`; resolve from `server.json` on ephemeral port |
| Daemon down | No meta / no daemon → warning stderr; exit 0; empty output |
| `--json` | Valid JSON array of snapshots; no table prose |
| Help | `session --help` / `fullHelp` lists `session list` |

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Host** (`RunDaemon`) binds the control HTTP server on an ephemeral port,
writes `{BaseDir}/server.json`, and serves `POST /v1/sessions` and `GET /v1/sessions`.

**SessionRegistry** holds live sessions. Each snapshot includes `session_id`, `phase`,
and `extension.connected`.

**Fake Extension** dials `GET /v1/ws?session=<id>`, sends `hello`, keeps the socket
open so the session becomes `extension_connected` with `connected: true`.

**Operator CLI** — `browser-agent session list [--base-dir] [--addr] [--json] [--color]`
resolves the control base URL, fetches sessions, and prints `FormatSessionList` human
output or raw JSON. When the daemon is unreachable, prints a stderr warning and empty
output without failing.

**Test Client** starts `RunDaemon`, creates sessions, optionally connects a fake
extension, invokes `HandleCLI session list`, and asserts stdout/stderr shape.

```text
RunDaemon(:0, BaseDir) -> server.json
POST /v1/sessions -> waiting_extension | extension_connected (with fake WS)

HandleCLI session list --base-dir BaseDir   # no --addr
  -> exit 0; table with Session ID / Phase / Connected; "N sessions"

HandleCLI session list --json --base-dir BaseDir
  -> exit 0; JSON array [{session_id, phase, extension: {connected}}]

no daemon / no server.json:
HandleCLI session list --base-dir BaseDir
  -> exit 0; stderr warning: daemon not running; stdout "0 sessions" or empty table

session --help -> lists session list
```

## Decision Tree

```
browser-agent-session-list
├── list/                                    [HandleCLI session list]
│   ├── empty/                                 daemon up, 0 sessions → exit 0; "0 sessions"
│   ├── two-sessions/                          create A+B → both ids in stdout, sorted
│   ├── phase-and-connected/                   waiting + fake WS hello → phase/connected cols
│   ├── from-server-json-no-addr/              ephemeral port; no --addr → lists correctly
│   ├── daemon-down-warning/                   no daemon → exit 0; stderr warning:
│   └── json-mode/                             --json → valid JSON array; no table prose
└── help/                                    [CLI help contract]
    └── mentions-list/                         session --help documents session list
```

### Parameter significance (high → low)

1. **Entry surface** — list CLI vs help.
2. **Daemon state** — running (live API) vs unreachable (warning path).
3. **Output mode** — human table vs `--json`.
4. **Session population** — empty vs multiple vs phase/connected mix.
5. **Addr source** — meta (`server.json`) vs explicit `--addr` (not in v1 leaves).

## Test Index

| Leaf | Scenario |
|------|----------|
| `list/empty` | Daemon up; no sessions; `session list` → exit 0; `0 sessions` or empty table |
| `list/two-sessions` | Create `sess-alpha` + `sess-zulu`; both ids in stdout sorted |
| `list/phase-and-connected` | One waiting + one fake WS hello; Phase and Connected columns correct |
| `list/from-server-json-no-addr` | Ephemeral port; omit `--addr`; lists sessions via `server.json` |
| `list/daemon-down-warning` | No daemon; exit 0; stderr `warning:` + daemon not running |
| `list/json-mode` | `--json` with one session → valid JSON array; no table headers |
| `help/mentions-list` | `session --help` contains `session list` or `list` under session |

**Leaf count: 7**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-list
doctest test ./tests/browser-agent-session-list    # RED until implementer lands session list
doctest test ./tests/browser-agent-daemon-phase3/v1-sessions/...
doctest test ./tests/browser-agent-session-delete/...
doctest test ./tests/browser-agent-session-addr-resolve/...
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until implementer
lands session list):

- `HandleCLI session list [--base-dir] [--addr] [--json] [--color] [--no-color]`
- `FormatSessionList(w, sessions []sessionSnapshot) error` (or equivalent)
- `cliSession` dispatch + `fullHelp` / `briefUsage` document `session list`
- Daemon down: stderr `warning: daemon not running in <base-dir>`; exit 0

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
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface under test.
const (
	ModeList = "list"
	ModeHelp = "help"
)

// ListOp — session list scenarios.
const (
	ListOpEmpty            = "empty"
	ListOpTwoSessions      = "two-sessions"
	ListOpPhaseConnected   = "phase-and-connected"
	ListOpFromServerJSON   = "from-server-json-no-addr"
	ListOpDaemonDown       = "daemon-down-warning"
	ListOpJSONMode         = "json-mode"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string
	ListOp string

	ModuleRoot string
	BaseDir    string
	Addr       string
	BaseURL    string

	StartDaemon bool
	OmitAddr    bool
	PassBaseDir bool
	JSONMode    bool

	SessionIDsToCreate []string
	ConnectExtensionFor string // session id to attach fake WS (phase-and-connected)

	HelloVersion  string
	HelloFeatures []string

	HelpArgs []string

	ReadyTimeout    time.Duration
	MaxDispatchWait time.Duration
	CLIEnv          map[string]string
}

// Response holds daemon + CLI list outcomes.
type Response struct {
	BaseURL string
	Addr    string

	CreatedSessionIDs []string
	WaitingSessionID  string
	ConnectedSessionID string

	Stdout   string
	Stderr   string
	ExitCode int
	CLIErr   string
	HelpText string

	DaemonMetaExists bool
	DaemonMeta       browseragent.DaemonMeta

	APISessionIDs []string
	APISessionsRaw string

	DispatchTimedOut bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	if req.MaxDispatchWait <= 0 {
		req.MaxDispatchWait = 12 * time.Second
	}

	switch req.Mode {
	case ModeList:
		return runListMode(t, req)
	case ModeHelp:
		return runHelpMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runListMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ListOp == "" {
		t.Fatal("ListOp must be set by leaf Setup")
	}

	resp := &Response{}

	if req.ListOp == ListOpDaemonDown {
		return runDaemonDownList(t, req, resp)
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return resp, err
	}
	defer cleanup()

	resp.BaseURL = srv.BaseURL
	resp.Addr = srv.Addr

	meta, _, exists, readErr := readDaemonMetaFile(req.BaseDir)
	resp.DaemonMetaExists = exists
	resp.DaemonMeta = meta
	if readErr != nil && !os.IsNotExist(readErr) {
		return resp, readErr
	}
	if !exists {
		return resp, fmt.Errorf("server.json missing under %s after RunDaemon", req.BaseDir)
	}

	if req.ListOp == ListOpFromServerJSON {
		if meta.Addr == "127.0.0.1:43761" || strings.HasSuffix(meta.Addr, ":43761") {
			return resp, fmt.Errorf("ephemeral daemon must not bind default :43761; got %q", meta.Addr)
		}
	}

	for _, sid := range req.SessionIDsToCreate {
		created, err := createSessionHTTP(srv.BaseURL, sid)
		if err != nil {
			return resp, err
		}
		resp.CreatedSessionIDs = append(resp.CreatedSessionIDs, created)
	}

	if req.ConnectExtensionFor != "" {
		ext, err := connectFakeExtension(t, srv.BaseURL, req.ConnectExtensionFor, req)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		time.Sleep(50 * time.Millisecond)
	}

	if len(resp.CreatedSessionIDs) >= 1 && req.ListOp == ListOpPhaseConnected {
		resp.WaitingSessionID = resp.CreatedSessionIDs[0]
		if req.ConnectExtensionFor != "" {
			resp.ConnectedSessionID = req.ConnectExtensionFor
		}
	}

	ids, raw, err := listSessionIDs(srv.BaseURL)
	if err != nil {
		return resp, err
	}
	resp.APISessionIDs = ids
	resp.APISessionsRaw = raw

	args := buildListArgs(req)
	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)
	return resp, err
}

func runDaemonDownList(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	metaPath := daemonMetaPath(req.BaseDir)
	_, statErr := os.Stat(metaPath)
	if statErr == nil {
		return resp, fmt.Errorf("daemon-down leaf requires absent server.json; found %s", metaPath)
	}
	if !os.IsNotExist(statErr) {
		return resp, statErr
	}
	resp.DaemonMetaExists = false

	args := buildListArgs(req)
	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)
	return resp, err
}

func runHelpMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	args := req.HelpArgs
	if len(args) == 0 {
		args = []string{"session", "--help"}
	}
	cliResp, err := invokeHandleCLI(t, req, args)
	resp := &Response{}
	mergeCLIResponse(resp, cliResp)
	resp.HelpText = resp.Stdout + resp.Stderr
	return resp, err
}

func buildListArgs(req *Request) []string {
	args := []string{"session", "list"}
	if req.PassBaseDir && req.BaseDir != "" {
		args = append(args, "--base-dir", req.BaseDir)
	}
	if !req.OmitAddr && req.BaseURL != "" {
		args = append(args, "--addr", req.BaseURL)
	}
	if req.JSONMode {
		args = append(args, "--json")
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

// --- RunDaemon harness (session-delete / session-addr-resolve) ---

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

func listSessionIDs(baseURL string) ([]string, string, error) {
	st, _, body, err := doGET(baseURL + "/v1/sessions")
	if err != nil {
		return nil, "", err
	}
	if st != http.StatusOK {
		return nil, string(body), fmt.Errorf("GET /v1/sessions status=%d body=%s", st, strings.TrimSpace(string(body)))
	}
	var snaps []map[string]any
	if err := json.Unmarshal(body, &snaps); err != nil {
		return nil, string(body), fmt.Errorf("parse GET /v1/sessions: %w", err)
	}
	ids := make([]string, 0, len(snaps))
	for _, snap := range snaps {
		if sid, ok := snap["session_id"].(string); ok && sid != "" {
			ids = append(ids, sid)
		}
	}
	sort.Strings(ids)
	return ids, string(body), nil
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

func connectFakeExtension(t *testing.T, baseURL, sessionID string, req *Request) (*fakeExtension, error) {
	t.Helper()
	ext, err := dialFakeExtension(baseURL, sessionID, req.HelloVersion, req.HelloFeatures)
	if err != nil {
		return nil, err
	}
	if err := ext.SendHello(); err != nil {
		ext.Close()
		return nil, err
	}
	go ext.Loop()
	return ext, nil
}

// --- fake extension WS client (phase4 harness) ---

type wsEnvelope struct {
	V       int            `json:"v"`
	Type    string         `json:"type"`
	ID      string         `json:"id"`
	Payload map[string]any `json:"payload"`
}

type fakeExtension struct {
	conn           *websocket.Conn
	version        string
	features       []string
	AutoCompleteOK bool
	OnJob          func(wsEnvelope)
	mu             sync.Mutex
	closed         bool
}

func dialFakeExtension(baseURL, sessionID, version string, features []string) (*fakeExtension, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
	if sessionID != "" {
		q := u.Query()
		q.Set("session", sessionID)
		u.RawQuery = q.Encode()
	}
	dialer := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
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
	return &fakeExtension{conn: conn, version: version, features: features}, nil
}

func (f *fakeExtension) SendHello() error {
	env := wsEnvelope{
		V:    1,
		Type: "hello",
		ID:   fmt.Sprintf("hello-%d", time.Now().UnixNano()),
		Payload: map[string]any{
			"version":            f.version,
			"features":           f.features,
			"browser_product":    "Chrome",
			"session_page_count": 1,
		},
	}
	return f.conn.WriteJSON(env)
}

func (f *fakeExtension) Loop() {
	for {
		f.mu.Lock()
		closed := f.closed
		f.mu.Unlock()
		if closed {
			return
		}
		var env wsEnvelope
		if err := f.conn.ReadJSON(&env); err != nil {
			return
		}
		switch env.Type {
		case "job":
			if f.OnJob != nil {
				f.OnJob(env)
			}
			if f.AutoCompleteOK {
				_ = f.sendResult(env, true, "", map[string]any{"ok": true})
			}
		case "ping":
			_ = f.conn.WriteJSON(wsEnvelope{V: 1, Type: "pong", ID: env.ID})
		}
	}
}

func (f *fakeExtension) sendResult(job wsEnvelope, ok bool, errMsg string, data map[string]any) error {
	jobID := job.ID
	if job.Payload != nil {
		if id, ok := job.Payload["id"].(string); ok && id != "" {
			jobID = id
		} else if id, ok := job.Payload["job_id"].(string); ok && id != "" {
			jobID = id
		}
	}
	env := wsEnvelope{
		V:    1,
		Type: "result",
		ID:   jobID,
		Payload: map[string]any{
			"job_id": jobID,
			"ok":     ok,
			"error":  errMsg,
			"data":   data,
		},
	}
	return f.conn.WriteJSON(env)
}

func (f *fakeExtension) Close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	_ = f.conn.Close()
}

var (
	_ = sort.Strings
	_ = sync.Mutex{}
	_ = io.Discard
)
```