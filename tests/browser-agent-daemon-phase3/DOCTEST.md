# browser-agent daemon Phase 3 — multi-session HTTP control plane

Refactors the control server to serve multiple sessions via **SessionRegistry**.
Adds session management HTTP endpoints and routes existing `/v1/session`, `/v1/jobs`,
and `/go` by `session` query/body. Asserts disconnected **hint** text and `/go`
**session warning** banner. One regression leaf uses `browseragent.Run` for
backward-compat single-session serve.

**No real Chrome.** **No real agent-run.** Integration via `httptest.Server` +
`SessionRegistry` + `NewRegistryControlHandler` (registry leaves). Fake extension
WS helpers are available for future leaves; phase 3 job leaf probes no-extension path.

| Surface | What is under test |
|---------|-------------------|
| `POST /v1/sessions` | Create via registry — 201, 409 duplicate, 400 invalid id |
| `GET /v1/sessions` | JSON array of session snapshots (sorted) |
| `GET /v1/session` | Required `?session=` in multi-session mode — 200/404/400 |
| `POST /v1/jobs` | Required `session_id` — 404 unknown; no-ext hint + session_url |
| `GET /go` | Session page + `data-browser-agent-session-warning` banner |
| Regression | `Run(ctx, cfg)` single session still answers `GET /v1/session?session=` |

Depends on Phase 1 helpers and Phase 2 `SessionRegistry`.

## Version

0.0.2

# DSN (Domain Specific Notion)

**SessionRegistry** (Phase 2) holds live `session` values keyed by id. **Registry
Control Server** is an HTTP mux backed by one registry — not the legacy
single-session `controlServer{sess}` only model.

**Test Client** builds a registry (temp `BaseDir`, loopback addr from
`httptest.Server`), optionally pre-creates sessions, mounts
`NewRegistryControlHandler(registry)`, and probes HTTP routes. **Fake Extension**
may dial `GET /v1/ws` (session routing is implementer detail for phase 3+).

### Session management

```text
POST /v1/sessions  body {session_id?} -> registry.Create -> 201 JSON | 409 | 400
GET  /v1/sessions  -> registry.List snapshots sorted by session_id
```

### Session-scoped routes (multi-session mode)

```text
GET  /v1/session?session=<id>   required param; empty/missing -> 400; unknown -> 404
POST /v1/jobs  body.session_id required; unknown -> 404
GET  /go?session=<id>           session SPA + data-browser-agent-session-warning
```

When `phase=waiting_extension` or extension not connected, snapshot `hint` and job
failure `data.hint` must mention keeping `/go?session=<id>` open and not navigating
to a different session in the same window.

### Backward compat

```text
Run(ctx, Config{SessionID}) -> registry.Create(SessionID) -> registry-backed server
GET /v1/session?session=<live id> -> 200 snapshot
```

## Decision Tree

```
browser-agent-daemon-phase3
├── v1-sessions/                         [session CRUD HTTP]
│   ├── post/
│   │   ├── create-201/                      POST body session_id -> 201 + JSON
│   │   ├── duplicate-409/                   second POST same id -> 409
│   │   └── invalid-400/                     invalid id -> 400
│   └── get/
│       ├── empty/                           fresh server -> []
│       └── two-sessions/                    two creates -> sorted array
├── v1-session/                          [GET snapshot by query]
│   ├── known-200/                           pre-created id + ?session= -> 200 + hint
│   ├── unknown-404/                         unknown ?session= -> 404
│   └── missing-param-400/                   no ?session= -> 400
├── v1-jobs/                             [POST jobs multi-session]
│   ├── known-session-no-ext-hint/           no WS -> ok:false + data.hint + session_url
│   └── unknown-404/                         unknown session_id -> 404
├── go-html/                             [GET /go session page]
│   ├── session-warning-banner/              warning marker + session id in HTML
│   └── unknown-session-404/                 unknown ?session= -> 404
└── regression-bridge/                   [Run() backward compat]
    └── run-single-session-still-works/      Run + GET /v1/session?session=id
```

### Parameter significance (high → low)

1. **HTTP surface** — sessions vs session vs jobs vs go vs regression (different routes).
2. **Outcome** — success vs client error (400/404/409) vs job failure with hint.
3. **Session presence** — known vs unknown vs missing param.
4. **Leaf specifics** — duplicate POST, sorted list, HTML marker text.

## Test Index

| Leaf | Scenario |
|------|----------|
| `v1-sessions/post/create-201` | POST `/v1/sessions` with session_id → 201 + create JSON |
| `v1-sessions/post/duplicate-409` | Second POST same session_id → 409 |
| `v1-sessions/post/invalid-400` | POST invalid session_id → 400 |
| `v1-sessions/get/empty` | GET `/v1/sessions` on fresh registry → `[]` |
| `v1-sessions/get/two-sessions` | Two sessions listed sorted ascending by id |
| `v1-session/known-200` | Known `?session=` → 200; waiting hint mentions `/go?session=` |
| `v1-session/unknown-404` | Unknown `?session=` → 404 |
| `v1-session/missing-param-400` | Missing `?session=` → 400 |
| `v1-jobs/known-session-no-ext-hint` | Known session, no extension → ok:false + disconnected hint |
| `v1-jobs/unknown-404` | Unknown session_id on POST → 404 |
| `go-html/session-warning-banner` | `/go?session=` HTML has warning banner + session id |
| `go-html/unknown-session-404` | `/go?session=unknown` → 404 |
| `regression-bridge/run-single-session-still-works` | `Run` single session; GET snapshot ok |

**Leaf count: 13**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase3
doctest test ./tests/browser-agent-daemon-phase3
# After implementer lands phase 3, regression elsewhere must stay green:
doctest test ./tests/browser-agent/...
```

Requires package `github.com/xhd2015/browser-agent/browseragent` exports (RED until
implementer):

- `NewSessionRegistry`, `NewRegistryControlHandler(registry *SessionRegistry) http.Handler`
- Multi-session routes on registry handler as specified above
- `Run(ctx, Config)` creates registry session and remains compatible with
  `tests/browser-agent/`

### Implementer contract (authoritative for GREEN)

**Registry HTTP**

| Method | Path | Notes |
|--------|------|-------|
| POST | `/v1/sessions` | body `{session_id?}`; 201 create result JSON; 409 `ErrSessionExists`; 400 invalid id |
| GET | `/v1/sessions` | JSON array of `sessionSnapshot`, sorted by `session_id` |
| GET | `/v1/session` | `?session=` **required**; 400 missing/empty; 404 unknown; 200 snapshot |
| POST | `/v1/jobs` | `session_id` **required** in body; 404 unknown; hold until result/timeout |
| GET | `/go` | `?session=` required for unknown lookup; HTML includes `data-browser-agent-session-warning` |
| GET | `/v1/health` | 200 liveness |
| GET | `/v1/ws` | extension WebSocket (per-session routing as needed) |

**Disconnected hint** (snapshot `hint` and job `data.hint` when not connected):

- Mention keeping session page open at `/go?session=<id>`
- Mention not closing or navigating to a different session in the same window

**Job failure without extension**: `ok:false`; error mentions not connected (or timeout
with hint); `data.hint` + `data.session_url` when possible.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — HTTP surface under test.
const (
	ModeV1SessionsPost = "v1-sessions-post"
	ModeV1SessionsGet  = "v1-sessions-get"
	ModeV1Session      = "v1-session"
	ModeV1Jobs         = "v1-jobs"
	ModeGoHTML         = "go-html"
	ModeRegression     = "regression-bridge"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	// Shared session ids
	SessionID             string
	PreCreateSessionIDs   []string
	SecondPreCreateID     string // two-sessions leaf: second id when set

	// POST /v1/sessions
	PostSessionID       string
	PostInvalidID       string
	PostDuplicateSecond bool

	// GET /v1/session
	SessionQueryParam   string
	OmitSessionQuery    bool
	ForceUnknownSession bool
	UnknownSessionID    string

	// POST /v1/jobs
	JobHTTPType        string
	JobHTTPParams      map[string]any
	JobHTTPTimeoutMS   int64
	JobSessionID       string
	OmitJobSessionID   bool

	// GET /go
	GoSessionQuery     string
	GoOmitSessionQuery bool
	GoUnknownSession   bool

	// regression Run()
	NoOpenChrome bool
	ReadyTimeout time.Duration
}

// Response holds HTTP outcomes for all modes.
type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string
	Raw         map[string]any

	// POST /v1/sessions
	CreatedSessionID  string
	CreatedSessionURL string
	SecondPostStatus  int

	// GET /v1/sessions
	SessionsListIDs []string

	// session snapshot
	SessionIDField       string
	Phase                string
	Hint                 string
	ExtensionConnected   bool
	SupportsBrowserAgent bool

	// jobs
	HTTPJobOK             bool
	HTTPJobError          string
	HTTPJobData           map[string]any
	HTTPJobDataHint       string
	HTTPJobDataSessionURL string

	// meta
	BaseURL       string
	RealSessionID string
	ProbeURL      string
	RunErrText    string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeV1SessionsPost:
		return runV1SessionsPost(t, req)
	case ModeV1SessionsGet:
		return runV1SessionsGet(t, req)
	case ModeV1Session:
		return runV1Session(t, req)
	case ModeV1Jobs:
		return runV1Jobs(t, req)
	case ModeGoHTML:
		return runGoHTML(t, req)
	case ModeRegression:
		return runRegressionBridge(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runV1SessionsPost(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL}

	body := map[string]any{}
	if req.PostInvalidID != "" {
		body["session_id"] = req.PostInvalidID
	} else if req.PostSessionID != "" {
		body["session_id"] = req.PostSessionID
	}

	status, ct, rawBody, err := postJSON(srv.BaseURL+"/v1/sessions", body)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = rawBody
	resp.BodyString = string(rawBody)
	parseCreateSessionJSON(resp, rawBody)

	if req.PostDuplicateSecond {
		status2, _, raw2, err := postJSON(srv.BaseURL+"/v1/sessions", body)
		if err != nil {
			return resp, err
		}
		resp.SecondPostStatus = status2
		_ = raw2
	}
	return resp, nil
}

func runV1SessionsGet(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// Pre-create via HTTP for list leaves that need sessions.
	for _, id := range req.PreCreateSessionIDs {
		st, _, b, perr := postJSON(srv.BaseURL+"/v1/sessions", map[string]any{"session_id": id})
		if perr != nil {
			return nil, perr
		}
		if st != http.StatusCreated {
			return nil, fmt.Errorf("pre-create %q: status %d body %s", id, st, string(b))
		}
	}
	if req.SecondPreCreateID != "" {
		st, _, b, perr := postJSON(srv.BaseURL+"/v1/sessions", map[string]any{"session_id": req.SecondPreCreateID})
		if perr != nil {
			return nil, perr
		}
		if st != http.StatusCreated {
			return nil, fmt.Errorf("pre-create %q: status %d body %s", req.SecondPreCreateID, st, string(b))
		}
	}

	status, ct, rawBody, err := doGET(srv.BaseURL + "/v1/sessions")
	if err != nil {
		return nil, err
	}
	resp := &Response{
		BaseURL:     srv.BaseURL,
		StatusCode:  status,
		ContentType: ct,
		Body:        rawBody,
		BodyString:  string(rawBody),
	}
	parseSessionsListJSON(resp, rawBody)
	return resp, nil
}

func runV1Session(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, RealSessionID: srv.PrimarySessionID}

	sessionID := srv.PrimarySessionID
	if req.ForceUnknownSession {
		sessionID = req.UnknownSessionID
		if sessionID == "" {
			sessionID = "does-not-exist"
		}
	}

	var probeURL string
	if req.OmitSessionQuery {
		probeURL = srv.BaseURL + "/v1/session"
	} else {
		probeURL = srv.BaseURL + "/v1/session?session=" + url.QueryEscape(sessionID)
	}
	status, ct, rawBody, err := doGET(probeURL)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = rawBody
	resp.BodyString = string(rawBody)
	resp.ProbeURL = probeURL
	parseSessionJSON(resp, rawBody)
	return resp, nil
}

func runV1Jobs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, RealSessionID: srv.PrimarySessionID}

	sessionForPost := srv.PrimarySessionID
	if req.ForceUnknownSession {
		sessionForPost = req.UnknownSessionID
		if sessionForPost == "" {
			sessionForPost = "does-not-exist"
		}
	}
	if req.JobSessionID != "" {
		sessionForPost = req.JobSessionID
	}

	timeoutMS := req.JobHTTPTimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 200
	}
	jobType := req.JobHTTPType
	if jobType == "" {
		jobType = "eval"
	}
	params := req.JobHTTPParams
	if params == nil {
		params = map[string]any{"code": "1+1"}
	}

	var status int
	var ct string
	var rawBody []byte
	var postErr error
	if req.OmitJobSessionID {
		status, ct, rawBody, postErr = postJobsNoSession(srv.BaseURL, jobType, params, timeoutMS)
	} else {
		status, ct, rawBody, postErr = postJobs(srv.BaseURL, sessionForPost, jobType, params, timeoutMS)
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = rawBody
	resp.BodyString = string(rawBody)
	resp.ProbeURL = srv.BaseURL + "/v1/jobs"
	if postErr != nil {
		return resp, postErr
	}
	parseJobHTTPResult(resp, rawBody)
	return resp, nil
}

func runGoHTML(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, RealSessionID: srv.PrimarySessionID}

	sessionID := srv.PrimarySessionID
	if req.GoUnknownSession {
		sessionID = req.UnknownSessionID
		if sessionID == "" {
			sessionID = "does-not-exist"
		}
	} else if req.GoSessionQuery != "" {
		sessionID = req.GoSessionQuery
	}

	var probeURL string
	if req.GoOmitSessionQuery {
		probeURL = srv.BaseURL + "/go"
	} else {
		probeURL = srv.BaseURL + "/go?session=" + url.QueryEscape(sessionID)
	}
	status, ct, rawBody, err := doGET(probeURL)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = rawBody
	resp.BodyString = string(rawBody)
	resp.ProbeURL = probeURL
	return resp, nil
}

func runRegressionBridge(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startAgentRunServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		RealSessionID: srv.SessionID,
		BaseURL:       srv.BaseURL,
	}
	if err := fillSessionProbe(resp, srv.BaseURL, srv.SessionID); err != nil {
		return resp, err
	}
	return resp, nil
}

// --- registry httptest harness ---

type registryHTTPServer struct {
	BaseURL          string
	PrimarySessionID string
	registry         *browseragent.SessionRegistry
	server           *httptest.Server
}

func startRegistryHTTPServer(t *testing.T, req *Request) (*registryHTTPServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by root Setup")
	}

	// Bind ephemeral port for addr metadata; httptest will use its own listener.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	if req.Addr != "" {
		addr = req.Addr
	}

	reg := browseragent.NewSessionRegistry(req.BaseDir, addr)
	for _, id := range req.PreCreateSessionIDs {
		if _, err := reg.Create(id); err != nil {
			return nil, nil, fmt.Errorf("registry pre-create %q: %w", id, err)
		}
	}
	if req.SecondPreCreateID != "" {
		if _, err := reg.Create(req.SecondPreCreateID); err != nil {
			return nil, nil, fmt.Errorf("registry pre-create %q: %w", req.SecondPreCreateID, err)
		}
	}

	handler := browseragent.NewRegistryControlHandler(reg)
	ts := httptest.NewServer(handler)
	baseURL := ts.URL

	primary := req.SessionID
	if primary == "" && len(req.PreCreateSessionIDs) > 0 {
		primary = req.PreCreateSessionIDs[0]
	}
	if primary == "" {
		primary = req.PostSessionID
	}

	out := &registryHTTPServer{
		BaseURL:          baseURL,
		PrimarySessionID: primary,
		registry:         reg,
		server:           ts,
	}
	cleanup := func() { ts.Close() }
	return out, cleanup, nil
}

// --- Run() regression harness (browser-agent pattern) ---

type agentServer struct {
	BaseURL   string
	SessionID string
	cancel    context.CancelFunc
	done      <-chan error
}

func startAgentRunServer(t *testing.T, req *Request) (*agentServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by root Setup")
	}
	if req.SessionID == "" {
		t.Fatal("SessionID must be set for regression Run")
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
		return nil, nil, fmt.Errorf("control server never healthy at %s: %w", baseURL, err)
	}

	srv := &agentServer{
		BaseURL:   baseURL,
		SessionID: req.SessionID,
		cancel:    cancel,
		done:      done,
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

// --- HTTP helpers ---

func postJSON(rawURL string, payload map[string]any) (int, string, []byte, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, "", nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(b))
	if err != nil {
		return 0, "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
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

func postJobs(baseURL, sessionID, jobType string, params map[string]any, timeoutMS int64) (int, string, []byte, error) {
	payload := map[string]any{
		"session_id": sessionID,
		"type":       jobType,
		"params":     params,
		"timeout_ms": timeoutMS,
	}
	return postJSON(baseURL+"/v1/jobs", payload)
}

func postJobsNoSession(baseURL, jobType string, params map[string]any, timeoutMS int64) (int, string, []byte, error) {
	payload := map[string]any{
		"type":       jobType,
		"params":     params,
		"timeout_ms": timeoutMS,
	}
	return postJSON(baseURL+"/v1/jobs", payload)
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
	u := baseURL + "/v1/session?session=" + url.QueryEscape(sessionID)
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

func parseCreateSessionJSON(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.Raw = raw
	if id, ok := raw["session_id"].(string); ok {
		resp.CreatedSessionID = id
	}
	if u, ok := raw["session_url"].(string); ok {
		resp.CreatedSessionURL = u
	}
}

func parseSessionsListJSON(resp *Response, body []byte) {
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err != nil {
		// Also accept {"sessions":[...]} wrapper.
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
	resp.Phase, _ = raw["phase"].(string)
	resp.Hint, _ = raw["hint"].(string)
	if ext, ok := raw["extension"].(map[string]any); ok {
		resp.ExtensionConnected, _ = ext["connected"].(bool)
		if v, ok := ext["supports_browser_agent"].(bool); ok {
			resp.SupportsBrowserAgent = v
		}
	}
}

func parseJobHTTPResult(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.Raw = raw
	src := raw
	if nested, ok := raw["result"].(map[string]any); ok {
		src = nested
	}
	if ok, exists := src["ok"].(bool); exists {
		resp.HTTPJobOK = ok
	}
	if e, ok := src["error"].(string); ok {
		resp.HTTPJobError = e
	}
	if resp.HTTPJobError == "" {
		if e, ok := raw["error"].(string); ok {
			resp.HTTPJobError = e
		}
	}
	if data, ok := src["data"].(map[string]any); ok {
		resp.HTTPJobData = data
		if h, ok := data["hint"].(string); ok {
			resp.HTTPJobDataHint = h
		}
		if u, ok := data["session_url"].(string); ok {
			resp.HTTPJobDataSessionURL = u
		}
	}
}

// --- fake extension WS client (reuse browser-agent harness pattern) ---

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
	JobsSeen       int
	mu             sync.Mutex
	closed         bool
}

func dialFakeExtension(baseURL, version string, features []string) (*fakeExtension, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
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
			f.mu.Lock()
			f.JobsSeen++
			f.mu.Unlock()
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