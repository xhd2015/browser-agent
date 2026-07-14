# browser-agent session rich — enriched snapshots, telemetry, human CLI

Classic TDD for richer **sessionSnapshot** fields, extension hello/status telemetry,
wider **session list** columns + delete hints, and **session info** human-default with
optional `--json`.

**No real Chrome.** **No real agent-run.** In-process `RunDaemon` on ephemeral
`127.0.0.1:0` with temp `BaseDir`; phase4-style fake extension WS for telemetry leaves.

| Surface | What is under test |
|---------|-------------------|
| `GET /v1/session` | `created_at`, `status`, `session_page_count`, `browsers`, derived status |
| WS hello telemetry | `session_page_count`, `browser_product`, `session_pages` in hello payload |
| WS status push | Tab open/close updates server page count |
| `session list` | Columns Created/Pages/Browser/Status; `—` for unknown; 0-page delete hint |
| `session info` | Human default sections; `--json` enriched machine output |

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Host** (`RunDaemon`) binds the control HTTP server, writes `{BaseDir}/server.json`,
and serves `GET /v1/session`, `GET /v1/sessions`, and `GET /v1/ws?session=<id>`.

**SessionRegistry** holds live sessions. Each **sessionSnapshot** is enriched with
`created_at`, optional `session_page_count` (nil = unknown → display `—`),
`browsers`, derived `status` + `status_label`, and optional detail fields.

**ComputeSessionStatus** derives status from count + extension connection:
count 0 → `no_session_page`; count >1 → `multiple_pages`; count 1 + disconnected →
`page_no_extension`; connected + supports BA → `ready`; connected + !supports →
`unsupported_extension`; count nil → `unknown`.

**Fake Extension** dials `GET /v1/ws?session=<id>`, sends `hello` with telemetry
(`browser_product`, `session_page_count`, `session_pages`), and may push `type=status`
when tab inventory changes.

**Operator CLI** — `session list` prints wider human columns; `session info` defaults
to human sections (Session, Created, Status, Pages, Browsers, Session URL, Next steps
with open-chrome / delete hints). `--json` emits enriched snapshot JSON.

**Test Client** starts `RunDaemon`, creates sessions, optionally connects fake extension
with telemetry, probes HTTP snapshots or invokes `HandleCLI`.

```text
RunDaemon(:0, BaseDir) -> server.json
POST /v1/sessions -> session id

Fake Extension -> hello { session_page_count, browser_product, session_pages }
GET /v1/session?session=ID -> created_at, status, session_page_count

Fake Extension -> status { session_page_count: 2 }
GET /v1/session -> session_page_count=2, status=multiple_pages

HandleCLI session list --base-dir BaseDir
  -> Session ID | Created | Pages | Browser | Status
  -> footer hint for 0-page sessions mentions session delete

HandleCLI session info --session-id ID --base-dir BaseDir
  -> human sections (NOT raw JSON-only)

HandleCLI session info --json ...
  -> JSON with created_at, status, session_page_count
```

## Decision Tree

```
browser-agent-session-rich
├── snapshot/                              [GET /v1/session API fields]
│   ├── created-at-in-json/                  created_at present
│   └── status-no-page/                       count=0 → status no_session_page
├── telemetry/                             [WS hello carries counts]
│   ├── hello-sets-page-count/               fake hello with session_page_count=1
│   └── hello-multi-page-orange/             count=2 → multiple_pages
├── list/                                  [session list human]
│   ├── columns-and-hint/                    columns Created/Pages/Browser/Status; 0-page hint mentions delete
│   └── unknown-pages-dash/                  no telemetry → Pages shows — (or unknown status)
├── info/                                  [session info human + json]
│   ├── human-disconnected/                  default human; mentions session URL + delete hint; NOT raw JSON-only
│   ├── human-ready/                         connected + hello; Status ready; human lines
│   └── json-flag-enriched/                  --json has created_at, status, session_page_count
└── extension/                             [WS status push updates count]
    └── status-push-updates-count/           tab status message updates server count
```

### Parameter significance (high → low)

1. **Surface** — API snapshot vs WS telemetry vs list CLI vs info CLI vs status push.
2. **Telemetry state** — unknown (nil) vs 0 pages vs 1 page vs multi-page.
3. **Extension connection** — disconnected vs connected + hello.
4. **Output mode** — human default vs `--json` (info only).

## Test Index

| Leaf | Scenario |
|------|----------|
| `snapshot/created-at-in-json` | `GET /v1/session` JSON includes non-empty `created_at` |
| `snapshot/status-no-page` | Hello with `session_page_count=0` → `status=no_session_page` |
| `telemetry/hello-sets-page-count` | Hello with count=1 → API `session_page_count=1`, `browser_product` |
| `telemetry/hello-multi-page-orange` | Hello with count=2 → `status=multiple_pages` |
| `list/columns-and-hint` | List columns Created/Pages/Browser/Status; 0-page footer hints delete |
| `list/unknown-pages-dash` | No telemetry → Pages column `—` and/or unknown status |
| `info/human-disconnected` | Default human output; session URL + delete hint; not JSON-only |
| `info/human-ready` | Connected hello count=1 → human Status ready |
| `info/json-flag-enriched` | `--json` includes `created_at`, `status`, `session_page_count` |
| `extension/status-push-updates-count` | Status push updates `session_page_count` on server |

**Leaf count: 10**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-rich
doctest test ./tests/browser-agent-session-rich    # RED until implementer lands session-rich
doctest test ./tests/browser-agent-session-list/...
doctest test ./tests/browser-agent-daemon-phase4/...
doctest test ./tests/browser-agent-product-hardening/session-info-cli/...
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until implementer
lands session-rich):

- Enriched `sessionSnapshot` + `ComputeSessionStatus`
- WS hello/status telemetry ingestion
- `FormatSessionList` wider columns + 0-page delete footer hint
- `session info` human default + `--json` enriched output

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
	ModeSnapshot  = "snapshot"
	ModeTelemetry = "telemetry"
	ModeList      = "list"
	ModeInfo      = "info"
	ModeExtension = "extension"
)

// SnapshotOp — GET /v1/session probes.
const (
	SnapshotOpCreatedAt   = "created-at-in-json"
	SnapshotOpStatusNoPage = "status-no-page"
)

// TelemetryOp — WS hello telemetry probes.
const (
	TelemetryOpHelloPageCount = "hello-sets-page-count"
	TelemetryOpMultiPage      = "hello-multi-page-orange"
)

// ListOp — session list human output.
const (
	ListOpColumnsHint    = "columns-and-hint"
	ListOpUnknownDash    = "unknown-pages-dash"
)

// InfoOp — session info CLI output.
const (
	InfoOpHumanDisconnected = "human-disconnected"
	InfoOpHumanReady        = "human-ready"
	InfoOpJSONEnriched      = "json-flag-enriched"
)

// ExtensionOp — WS status push.
const (
	ExtensionOpStatusPush = "status-push-updates-count"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	SnapshotOp  string
	TelemetryOp string
	ListOp      string
	InfoOp      string
	ExtensionOp string

	ModuleRoot string
	BaseDir    string
	Addr       string
	BaseURL    string

	SessionID string

	ConnectExtension bool
	SendHello        bool
	HelloVersion     string
	HelloFeatures    []string

	// Telemetry payload (hello / status)
	BrowserProduct    string
	SessionPageCount  *int
	SessionPages      []map[string]any
	StatusPushCount   int
	StatusPushPages   []map[string]any

	JSONMode    bool
	PassBaseDir bool
	OmitAddr    bool

	ReadyTimeout    time.Duration
	MaxDispatchWait time.Duration
	CLIEnv          map[string]string
}

// Response holds daemon + HTTP + CLI outcomes.
type Response struct {
	BaseURL   string
	Addr      string
	SessionID string

	// HTTP snapshot
	SessionStatusCode int
	SessionBody       []byte
	SessionBodyString string
	SessionJSON       map[string]any

	CreatedAt         string
	Status            string
	StatusLabel       string
	SessionPageCount  *int
	Browsers          []string

	// WS
	WSHelloOK    bool
	WSStatusSent bool

	// CLI
	Stdout   string
	Stderr   string
	ExitCode int
	CLIErr   string

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
	case ModeSnapshot:
		return runSnapshotMode(t, req)
	case ModeTelemetry:
		return runTelemetryMode(t, req)
	case ModeList:
		return runListMode(t, req)
	case ModeInfo:
		return runInfoMode(t, req)
	case ModeExtension:
		return runExtensionMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runSnapshotMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SnapshotOp == "" {
		t.Fatal("SnapshotOp must be set by leaf Setup")
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
	req.SessionID = sid

	if req.SnapshotOp == SnapshotOpStatusNoPage {
		zero := 0
		req.SessionPageCount = &zero
		req.ConnectExtension = true
		req.SendHello = true
		ext, err := connectFakeExtensionTelemetry(t, srv.BaseURL, sid, req)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		resp.WSHelloOK = true
		time.Sleep(60 * time.Millisecond)
	}

	if err := fillSessionProbe(resp, srv.BaseURL, sid); err != nil {
		return resp, err
	}
	return resp, nil
}

func runTelemetryMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.TelemetryOp == "" {
		t.Fatal("TelemetryOp must be set by leaf Setup")
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
	req.SessionID = sid

	switch req.TelemetryOp {
	case TelemetryOpHelloPageCount:
		one := 1
		req.SessionPageCount = &one
		req.BrowserProduct = "Chrome"
		req.SessionPages = []map[string]any{
			{"tab_id": 1, "url": "http://127.0.0.1:43761/go?session=" + sid},
		}
	case TelemetryOpMultiPage:
		two := 2
		req.SessionPageCount = &two
		req.BrowserProduct = "Chrome"
		req.SessionPages = []map[string]any{
			{"tab_id": 1, "url": "http://127.0.0.1:43761/go?session=" + sid},
			{"tab_id": 2, "url": "https://example.com/"},
		}
	}

	req.ConnectExtension = true
	req.SendHello = true
	ext, err := connectFakeExtensionTelemetry(t, srv.BaseURL, sid, req)
	if err != nil {
		return resp, err
	}
	defer ext.Close()
	resp.WSHelloOK = true
	time.Sleep(60 * time.Millisecond)

	if err := fillSessionProbe(resp, srv.BaseURL, sid); err != nil {
		return resp, err
	}
	return resp, nil
}

func runListMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ListOp == "" {
		t.Fatal("ListOp must be set by leaf Setup")
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
	req.SessionID = sid

	switch req.ListOp {
	case ListOpColumnsHint:
		zero := 0
		req.SessionPageCount = &zero
		req.ConnectExtension = true
		req.SendHello = true
		ext, err := connectFakeExtensionTelemetry(t, srv.BaseURL, sid, req)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		time.Sleep(60 * time.Millisecond)
	case ListOpUnknownDash:
		// No extension / no telemetry — snapshot page count stays unknown.
	}

	args := []string{"session", "list", "--base-dir", req.BaseDir}
	if !req.OmitAddr && srv.BaseURL != "" {
		args = append(args, "--addr", srv.BaseURL)
	}
	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)
	return resp, err
}

func runInfoMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.InfoOp == "" {
		t.Fatal("InfoOp must be set by leaf Setup")
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
	req.SessionID = sid

	if req.InfoOp == InfoOpHumanReady || req.InfoOp == InfoOpJSONEnriched {
		one := 1
		req.SessionPageCount = &one
		req.BrowserProduct = "Chrome"
		req.ConnectExtension = true
		req.SendHello = true
		ext, err := connectFakeExtensionTelemetry(t, srv.BaseURL, sid, req)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		// session info enqueues type=info when connected — auto-complete so CLI returns.
		ext.AutoCompleteOK = true
		ext.ResultData = map[string]any{
			"tabs": []map[string]any{
				{"id": 1, "url": "http://127.0.0.1:43761/go?session=" + sid, "title": "Session"},
			},
			"version":  req.HelloVersion,
			"features": []any{"browser-agent"},
		}
		time.Sleep(60 * time.Millisecond)
	}

	args := []string{
		"session", "info",
		"--session-id", sid,
		"--base-dir", req.BaseDir,
	}
	if !req.OmitAddr && srv.BaseURL != "" {
		args = append(args, "--addr", srv.BaseURL)
	}
	if req.JSONMode || req.InfoOp == InfoOpJSONEnriched {
		args = append(args, "--json")
	}

	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)
	if err != nil {
		return resp, err
	}

	if req.InfoOp == InfoOpJSONEnriched {
		parseStdoutSnapshotFields(resp)
	}
	return resp, nil
}

func runExtensionMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ExtensionOp == "" {
		t.Fatal("ExtensionOp must be set by leaf Setup")
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
	req.SessionID = sid

	one := 1
	req.SessionPageCount = &one
	req.BrowserProduct = "Chrome"
	req.ConnectExtension = true
	req.SendHello = true
	ext, err := connectFakeExtensionTelemetry(t, srv.BaseURL, sid, req)
	if err != nil {
		return resp, err
	}
	defer ext.Close()
	resp.WSHelloOK = true
	time.Sleep(60 * time.Millisecond)

	if err := fillSessionProbe(resp, srv.BaseURL, sid); err != nil {
		return resp, err
	}
	before := -1
	if resp.SessionPageCount != nil {
		before = *resp.SessionPageCount
	}

	pushCount := req.StatusPushCount
	if pushCount <= 0 {
		pushCount = 2
	}
	pushPages := req.StatusPushPages
	if pushPages == nil {
		pushPages = []map[string]any{
			{"tab_id": 1, "url": "http://127.0.0.1:43761/go?session=" + sid},
			{"tab_id": 2, "url": "https://example.com/"},
		}
	}
	if err := ext.SendStatus(req.BrowserProduct, pushCount, pushPages); err != nil {
		return resp, err
	}
	resp.WSStatusSent = true
	time.Sleep(60 * time.Millisecond)

	if err := fillSessionProbe(resp, srv.BaseURL, sid); err != nil {
		return resp, err
	}
	_ = before
	return resp, nil
}

func fillSessionProbe(resp *Response, baseURL, sessionID string) error {
	st, _, body, err := doGET(baseURL + "/v1/session?session=" + url.QueryEscape(sessionID))
	if err != nil {
		return err
	}
	resp.SessionStatusCode = st
	resp.SessionBody = body
	resp.SessionBodyString = string(body)
	parseSnapshotBody(resp, body)
	return nil
}

func parseSnapshotBody(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.SessionJSON = raw
	if v, ok := raw["created_at"].(string); ok {
		resp.CreatedAt = v
	}
	if v, ok := raw["status"].(string); ok {
		resp.Status = v
	}
	if v, ok := raw["status_label"].(string); ok {
		resp.StatusLabel = v
	}
	if v, ok := raw["session_page_count"].(float64); ok {
		n := int(v)
		resp.SessionPageCount = &n
	}
	if arr, ok := raw["browsers"].([]any); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				resp.Browsers = append(resp.Browsers, s)
			}
		}
	}
}

func parseStdoutSnapshotFields(resp *Response) {
	if resp == nil || strings.TrimSpace(resp.Stdout) == "" {
		return
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Stdout)), &raw); err != nil {
		return
	}
	resp.SessionJSON = raw
	if v, ok := raw["created_at"].(string); ok {
		resp.CreatedAt = v
	}
	if v, ok := raw["status"].(string); ok {
		resp.Status = v
	}
	if v, ok := raw["session_page_count"].(float64); ok {
		n := int(v)
		resp.SessionPageCount = &n
	}
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

// --- RunDaemon harness (session-list / session-delete pattern) ---

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

func connectFakeExtensionTelemetry(t *testing.T, baseURL, sessionID string, req *Request) (*fakeExtension, error) {
	t.Helper()
	ext, err := dialFakeExtension(baseURL, sessionID, req.HelloVersion, req.HelloFeatures)
	if err != nil {
		return nil, err
	}
	if req.SendHello {
		if err := ext.SendHelloTelemetry(req.BrowserProduct, req.SessionPageCount, req.SessionPages); err != nil {
			ext.Close()
			return nil, err
		}
	}
	go ext.Loop()
	return ext, nil
}

// --- fake extension WS client (phase4 harness + telemetry) ---

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
	ResultData     map[string]any
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

func (f *fakeExtension) SendHelloTelemetry(browserProduct string, pageCount *int, pages []map[string]any) error {
	payload := map[string]any{
		"version":  f.version,
		"features": f.features,
	}
	if browserProduct != "" {
		payload["browser_product"] = browserProduct
	}
	if pageCount != nil {
		payload["session_page_count"] = *pageCount
	}
	if pages != nil {
		payload["session_pages"] = pages
	}
	env := wsEnvelope{
		V:       1,
		Type:    "hello",
		ID:      fmt.Sprintf("hello-%d", time.Now().UnixNano()),
		Payload: payload,
	}
	return f.conn.WriteJSON(env)
}

func (f *fakeExtension) SendStatus(browserProduct string, pageCount int, pages []map[string]any) error {
	payload := map[string]any{
		"session_page_count": pageCount,
	}
	if browserProduct != "" {
		payload["browser_product"] = browserProduct
	}
	if pages != nil {
		payload["session_pages"] = pages
	}
	env := wsEnvelope{
		V:       1,
		Type:    "status",
		ID:      fmt.Sprintf("status-%d", time.Now().UnixNano()),
		Payload: payload,
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
				data := f.ResultData
				if data == nil {
					data = map[string]any{"ok": true}
				}
				_ = f.sendResult(env, true, "", data)
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