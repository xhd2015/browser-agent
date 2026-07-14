# browser-agent session delete — CLI + HTTP + disk-only cleanup

Greenfield feature: **`browser-agent session delete`** and **`DELETE /v1/session?session=<id>`**
remove sessions from the registry, fail inflight jobs, and `RemoveAll` the session directory.

**Reject delete when extension connected** (`isExtensionConnected()` / live WS).
**Allow delete when `waiting_extension`** (disconnected). **Disk-only dirs** (on disk, not in
registry) → cleanup delete allowed.

**No real Chrome.** **No real agent-run.** In-process `RunDaemon` on ephemeral
`127.0.0.1:0` with temp `BaseDir`; fake extension WS hello for connected rejection leaves.

| Surface | What is under test |
|---------|-------------------|
| `session delete` CLI | Omit `--addr`; resolve from `server.json`; stdout deleted message |
| `DELETE /v1/session` | 200/204 when waiting; 409 when extension connected |
| Disk-only cleanup | Dir on disk without registry entry → delete removes dir |
| Help | `session --help` / `fullHelp` lists `session delete` |

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Host** (`RunDaemon`) binds the control HTTP server on an ephemeral port,
writes `{BaseDir}/server.json`, and serves session lifecycle routes including
`POST /v1/sessions`, `GET /v1/sessions`, and (after implementer) `DELETE /v1/session`.

**SessionRegistry** holds live sessions. **`Delete(id)`** removes the registry
entry, calls `FailAllInflight` on the session job queue, and `os.RemoveAll` on
`{baseDir}/sessions/{id}/`. Rejects when `isExtensionConnected()` is true.
Disk-only ids (dir exists, not in map) may delete the directory only.

**Fake Extension** dials `GET /v1/ws?session=<id>`, sends `hello`, keeps the
socket open so the session is `extension_connected`.

**Operator CLI** — `browser-agent session delete --session-id <id> [--base-dir] [--addr]`
invokes delete against the resolved control base URL. Success prints
`deleted session <id>`; failures surface not-found or extension-connected errors.

**Test Client** starts `RunDaemon`, creates or seeds sessions, optionally connects
a fake extension, invokes CLI or HTTP delete, and asserts registry/list/dir state.

```text
RunDaemon(:0, BaseDir) -> server.json
POST /v1/sessions -> waiting_extension

HandleCLI session delete --session-id ID --base-dir BaseDir
  -> exit 0; deleted message; SessionDirExists false; absent from GET /v1/sessions

Fake Extension -> hello on /v1/ws?session=ID
HandleCLI session delete ... -> exit 1; extension connected; session remains

DELETE /v1/session?session=ID -> 200|204 (waiting) or 409 (connected)

MkdirAll sessions/id (disk-only) -> delete -> dir gone; exit 0
session --help -> lists session delete
```

## Decision Tree

```
browser-agent-session-delete
├── cli/                                   [HandleCLI session delete]
│   ├── waiting-extension-ok/                create session, no WS; delete → exit 0; dir gone
│   ├── connected-rejected/                  fake WS hello; delete → exit 1; extension connected
│   └── not-found/                           unknown id → exit 1; not found
├── http/                                  [DELETE /v1/session API]
│   ├── waiting-extension-202/               DELETE → 200 or 204; session gone from GET list
│   └── connected-409/                       hello then DELETE → 409
├── disk-only/                             [cleanup without registry entry]
│   └── cleanup-ok/                          mkdir sessions/id on disk only; delete → dir gone
└── help/                                  [CLI help contract]
    └── mentions-delete/                     session --help lists session delete
```

### Parameter significance (high → low)

1. **Entry surface** — CLI vs HTTP vs disk-only vs help.
2. **Extension state** — waiting_extension (allow) vs extension_connected (reject).
3. **Session presence** — registered session vs disk-only vs not found.

## Test Index

| Leaf | Scenario |
|------|----------|
| `cli/waiting-extension-ok` | Create session; no WS; `session delete` → exit 0; dir gone; not in list |
| `cli/connected-rejected` | Fake hello; delete → exit 1; extension connected in error; session remains |
| `cli/not-found` | Unknown id → exit 1; not found message |
| `http/waiting-extension-202` | DELETE → 200 or 204; id absent from `GET /v1/sessions` |
| `http/connected-409` | Hello then DELETE → 409; body mentions extension connected |
| `disk-only/cleanup-ok` | Disk-only dir; delete → dir gone; exit 0 |
| `help/mentions-delete` | `session --help` contains `session delete` or `delete` under session |

**Leaf count: 7**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-delete
doctest test ./tests/browser-agent-session-delete    # RED after design
doctest test ./tests/browser-agent-daemon-phase4/... # WS regressions
doctest test ./tests/browser-agent-session-addr-resolve/...
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until implementer
lands session delete):

- `SessionRegistry.Delete(id string) error`
- `DELETE /v1/session?session=<id>` on registry control handler
- `HandleCLI session delete --session-id <id> [--base-dir] [--addr]`
- `fullHelp` / `briefUsage` document `session delete`

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

// Mode — top-level surface under test.
const (
	ModeCLI      = "cli"
	ModeHTTP     = "http"
	ModeDiskOnly = "disk-only"
	ModeHelp     = "help"
)

// CLIOp — CLI delete scenarios.
const (
	CLIOpWaitingOK      = "waiting-extension-ok"
	CLIOpConnectedReject = "connected-rejected"
	CLIOpNotFound       = "not-found"
)

// HTTPOp — HTTP DELETE scenarios.
const (
	HTTPOpWaiting202   = "waiting-extension-202"
	HTTPOpConnected409 = "connected-409"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string
	BaseURL    string

	CLIOp  string
	HTTPOp string

	SessionID         string
	UnknownSessionID  string
	DiskOnlySessionID string

	ConnectExtension bool
	HelloVersion     string
	HelloFeatures    []string

	HelpArgs []string

	ReadyTimeout    time.Duration
	MaxDispatchWait time.Duration
	CLIEnv          map[string]string
}

// Response holds daemon + CLI + HTTP delete outcomes.
type Response struct {
	BaseURL string
	Addr    string

	SessionID string

	Stdout   string
	Stderr   string
	ExitCode int
	CLIErr   string
	HelpText string

	DeleteStatusCode int
	DeleteBodyString string

	SessionDirExists bool
	SessionInList    bool
	SessionsListRaw  string

	ExtensionConnectedBeforeDelete bool
	ExtensionConnectedAfterReject  bool

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
	case ModeCLI:
		return runCLIMode(t, req)
	case ModeHTTP:
		return runHTTPMode(t, req)
	case ModeDiskOnly:
		return runDiskOnlyMode(t, req)
	case ModeHelp:
		return runHelpMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runCLIMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLIOp == "" {
		t.Fatal("CLIOp must be set by leaf Setup")
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, Addr: srv.Addr}

	var targetID string
	switch req.CLIOp {
	case CLIOpNotFound:
		targetID = req.UnknownSessionID
		if targetID == "" {
			targetID = "sess-does-not-exist"
		}
	default:
		sid, err := createSessionHTTP(srv.BaseURL, req.SessionID)
		if err != nil {
			return resp, err
		}
		targetID = sid
		resp.SessionID = sid
		req.SessionID = sid
	}

	var ext *fakeExtension
	if req.ConnectExtension && req.CLIOp == CLIOpConnectedReject {
		ext, err = connectFakeExtension(t, srv.BaseURL, targetID, req)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		time.Sleep(50 * time.Millisecond)
		connected, err := probeExtensionConnected(srv.BaseURL, targetID)
		if err != nil {
			return resp, err
		}
		resp.ExtensionConnectedBeforeDelete = connected
	}

	args := []string{
		"session", "delete",
		"--session-id", targetID,
		"--base-dir", req.BaseDir,
	}
	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)
	if err != nil {
		return resp, err
	}

	if req.CLIOp == CLIOpWaitingOK {
		resp.SessionDirExists = browseragent.SessionDirExists(req.BaseDir, targetID)
		ids, listRaw, err := listSessionIDs(srv.BaseURL)
		if err != nil {
			return resp, err
		}
		resp.SessionsListRaw = listRaw
		resp.SessionInList = containsID(ids, targetID)
	}

	if req.CLIOp == CLIOpConnectedReject {
		resp.SessionDirExists = browseragent.SessionDirExists(req.BaseDir, targetID)
		ids, listRaw, err := listSessionIDs(srv.BaseURL)
		if err != nil {
			return resp, err
		}
		resp.SessionsListRaw = listRaw
		resp.SessionInList = containsID(ids, targetID)
		connected, err := probeExtensionConnected(srv.BaseURL, targetID)
		if err != nil {
			return resp, err
		}
		resp.ExtensionConnectedAfterReject = connected
	}

	return resp, nil
}

func runHTTPMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.HTTPOp == "" {
		t.Fatal("HTTPOp must be set by leaf Setup")
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, Addr: srv.Addr}

	sid, err := createSessionHTTP(srv.BaseURL, req.SessionID)
	if err != nil {
		return resp, err
	}
	resp.SessionID = sid

	var ext *fakeExtension
	if req.ConnectExtension && req.HTTPOp == HTTPOpConnected409 {
		ext, err = connectFakeExtension(t, srv.BaseURL, sid, req)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		time.Sleep(50 * time.Millisecond)
	}

	delURL := srv.BaseURL + "/v1/session?session=" + url.QueryEscape(sid)
	status, body, err := doDELETE(delURL)
	if err != nil {
		return resp, err
	}
	resp.DeleteStatusCode = status
	resp.DeleteBodyString = string(body)

	if req.HTTPOp == HTTPOpWaiting202 {
		resp.SessionDirExists = browseragent.SessionDirExists(req.BaseDir, sid)
		ids, listRaw, err := listSessionIDs(srv.BaseURL)
		if err != nil {
			return resp, err
		}
		resp.SessionsListRaw = listRaw
		resp.SessionInList = containsID(ids, sid)
	}

	return resp, nil
}

func runDiskOnlyMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, Addr: srv.Addr}

	diskID := req.DiskOnlySessionID
	if diskID == "" {
		diskID = "sess-disk-only-cleanup"
	}
	resp.SessionID = diskID

	sessionDir := browseragent.SessionDirPath(req.BaseDir, diskID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return resp, err
	}
	if !browseragent.SessionDirExists(req.BaseDir, diskID) {
		return resp, fmt.Errorf("seeded disk-only dir missing at %s", sessionDir)
	}

	args := []string{
		"session", "delete",
		"--session-id", diskID,
		"--base-dir", req.BaseDir,
	}
	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)
	if err != nil {
		return resp, err
	}

	resp.SessionDirExists = browseragent.SessionDirExists(req.BaseDir, diskID)
	return resp, nil
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

func probeExtensionConnected(baseURL, sessionID string) (bool, error) {
	u := baseURL + "/v1/session?session=" + url.QueryEscape(sessionID)
	st, _, body, err := doGET(u)
	if err != nil {
		return false, err
	}
	if st != http.StatusOK {
		return false, fmt.Errorf("GET %s status=%d body=%s", u, st, strings.TrimSpace(string(body)))
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return false, err
	}
	if ext, ok := raw["extension"].(map[string]any); ok {
		connected, _ := ext["connected"].(bool)
		return connected, nil
	}
	return false, nil
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
	return ids, string(body), nil
}

func containsID(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
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

// --- RunDaemon harness (phase6 / session-addr-resolve) ---

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

func doDELETE(rawURL string) (int, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, rawURL, nil)
	if err != nil {
		return 0, nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, err
	}
	return res.StatusCode, body, nil
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
			"version":  f.version,
			"features": f.features,
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
	_ = sync.Mutex{}
	_ = io.Discard
)
```