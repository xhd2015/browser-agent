# browser-agent daemon Phase 4 — per-session WebSocket isolation

Hardens the registry-backed WebSocket layer for **true multi-session isolation**:
required `?session=` when two or more sessions exist, hello/disconnect/job routing
scoped per session socket.

**No real Chrome.** **No real agent-run.** Integration via `httptest.Server` +
`SessionRegistry` + `NewRegistryControlHandler` + gorilla websocket fake extension
(reuses browser-agent WS harness patterns).

| Surface | What is under test |
|---------|-------------------|
| `GET /v1/ws?session=` | Required when registry has 2+ sessions — 400 if missing; known id upgrades |
| Hello isolation | Hello on session A → only A `extension.connected=true`; B stays false |
| Disconnect isolation | Close A WS → only A inflight job fails; B stays connected and can run jobs |
| Job routing | POST job on A → `type=job` only on A's socket; B socket must not receive |

Depends on Phases 1–3 (`SessionRegistry`, `NewRegistryControlHandler`). Keep
`tests/browser-agent/ws-control/...` GREEN (single-session `Run` backward compat).

## Version

0.0.2

# DSN (Domain Specific Notion)

**SessionRegistry** holds multiple live **session** values. **Registry Control
Server** routes `GET /v1/ws?session=<id>` to the matching session's extension
socket. When the registry has **two or more** sessions, the `session` query param
is **required** — missing param → **400**. With exactly **one** session, missing
param may fall back to that id (backward compat for `Run` single-session serve;
not exercised in this tree).

**Fake Extension** dials `GET /v1/ws?session=<id>`, sends `hello`, receives
`type=job` pushes, may send `type=result`. **Disconnect policy v1**: WS drop
fails inflight jobs on **that session only**; other sessions unaffected.

**Test Client** builds registry + `httptest.Server`, pre-creates sessions A and B,
dials one or two fake extensions, probes `GET /v1/session?session=` snapshots and
`POST /v1/jobs` with `session_id`.

```text
GET /v1/ws?session=A     -> upgrade -> hello -> session A connected
GET /v1/ws (2+ sessions) -> 400 missing session
POST /v1/jobs session_id=A -> push type=job on A socket only
close A WS -> A inflight jobs fail; B socket + jobs unaffected
```

## Decision Tree

```
browser-agent-daemon-phase4
├── ws-session-param/                    [WS dial session query]
│   ├── two-sessions-missing-400/          2 sessions, /v1/ws no param → 400
│   └── two-sessions-known-upgrade/        /v1/ws?session=A → upgrade ok
├── ws-hello-isolation/                  [hello scopes connection state]
│   ├── hello-a-only-a-connected/          hello on A; A connected, B false
│   └── hello-b-only-b-connected/          hello on B; B connected, A false
├── ws-disconnect-isolation/             [disconnect scoped per session]
│   └── disconnect-a-inflight-only-a/      inflight on A; close A; B ok + job
└── ws-job-routing/                      [job push scoped per socket]
    └── job-push-session-a-only/           job on A; only A socket receives
```

### Parameter significance (high → low)

1. **WS surface** — session param vs hello vs disconnect vs job routing.
2. **Target session** — A vs B (which socket / snapshot is probed).
3. **Outcome** — 400 vs upgrade vs connected isolation vs job delivery.

## Test Index

| Leaf | Scenario |
|------|----------|
| `ws-session-param/two-sessions-missing-400` | 2 sessions; `GET /v1/ws` without `session` → 400 |
| `ws-session-param/two-sessions-known-upgrade` | 2 sessions; `GET /v1/ws?session=A` → WebSocket upgrade ok |
| `ws-hello-isolation/hello-a-only-a-connected` | Hello on A; A connected+supports; B disconnected |
| `ws-hello-isolation/hello-b-only-b-connected` | Hello on B; B connected+supports; A disconnected |
| `ws-disconnect-isolation/disconnect-a-inflight-only-a` | Inflight job on A; close A WS → A fails; B connected + job ok |
| `ws-job-routing/job-push-session-a-only` | WS on A and B; POST job A → only A receives `type=job` |

**Leaf count: 6**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase4
doctest test ./tests/browser-agent-daemon-phase4
# After implementer lands phase 4:
doctest test ./tests/browser-agent/ws-control/...
doctest test ./tests/browser-agent/...
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until
implementer hardens per-session WS routing):

- `GET /v1/ws?session=` required when registry has **2+** sessions (400 if missing)
- Per-session hello updates only that session's `extension.connected`
- Per-session disconnect fails only that session's inflight jobs
- `pushJob` delivers only on the matching session's active socket

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

// Mode — WS surface under test.
const (
	ModeWSSessionParam        = "ws-session-param"
	ModeWSHelloIsolation      = "ws-hello-isolation"
	ModeWSDisconnectIsolation = "ws-disconnect-isolation"
	ModeWSJobRouting          = "ws-job-routing"
)

// WSSessionOp — session param dial probes.
const (
	WSSessionOpMissing400    = "missing-400"
	WSSessionOpKnownUpgrade  = "known-upgrade"
)

// WSHelloTarget — which session receives hello.
const (
	WSHelloTargetA = "A"
	WSHelloTargetB = "B"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	// Two-session fixtures (defaults set in root Setup when empty).
	SessionIDA string
	SessionIDB string

	PreCreateSessionIDs []string
	SecondPreCreateID   string

	// ws-session-param
	WSSessionOp      string
	OmitWSSessionParam bool
	DialSessionID    string

	// ws-hello-isolation
	WSHelloTarget string

	// shared WS / jobs
	HelloVersion  string
	HelloFeatures []string
	JobHTTPType   string
	JobHTTPParams map[string]any
	JobHTTPTimeoutMS int64
	JobTargetSessionID string

	ReadyTimeout time.Duration
}

// Response holds WS + HTTP outcomes.
type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string
	Raw         map[string]any

	// WS dial
	WSDialOK     bool
	WSDialStatus int
	WSDialErr    string

	// hello / snapshots (session A and B)
	WSHelloOK            bool
	ExtensionConnectedA  bool
	ExtensionConnectedB  bool
	SupportsA            bool
	SupportsB            bool
	SessionAProbeURL     string
	SessionBProbeURL     string

	// jobs / disconnect
	WSJobReceivedOnA bool
	WSJobReceivedOnB bool
	WSJobType        string
	WSJobPayloadRaw  string
	WSDisconnected   bool
	HTTPJobOK        bool
	HTTPJobError     string
	HTTPJobOKOnB     bool
	HTTPJobErrorOnB  string

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
	case ModeWSSessionParam:
		return runWSSessionParam(t, req)
	case ModeWSHelloIsolation:
		return runWSHelloIsolation(t, req)
	case ModeWSDisconnectIsolation:
		return runWSDisconnectIsolation(t, req)
	case ModeWSJobRouting:
		return runWSJobRouting(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runWSSessionParam(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.WSSessionOp == "" {
		t.Fatal("WSSessionOp must be set by leaf Setup")
	}
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL}

	dialID := req.DialSessionID
	if req.WSSessionOp == WSSessionOpMissing400 || req.OmitWSSessionParam {
		dialID = ""
	}
	if dialID == "" && req.WSSessionOp == WSSessionOpKnownUpgrade {
		dialID = req.SessionIDA
	}

	status, dialErr := probeWSDial(srv.BaseURL, dialID)
	resp.WSDialStatus = status
	if dialErr != nil {
		resp.WSDialErr = dialErr.Error()
	}
	resp.WSDialOK = dialErr == nil
	resp.StatusCode = status
	return resp, nil
}

func runWSHelloIsolation(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.WSHelloTarget == "" {
		t.Fatal("WSHelloTarget must be set by leaf Setup")
	}
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL}

	version := req.HelloVersion
	if version == "" {
		version = "1.0.0"
	}
	features := req.HelloFeatures
	if features == nil {
		features = []string{"browser-agent"}
	}

	targetID := req.SessionIDA
	if req.WSHelloTarget == WSHelloTargetB {
		targetID = req.SessionIDB
	}

	ext, err := dialFakeExtension(srv.BaseURL, targetID, version, features)
	if err != nil {
		return resp, err
	}
	defer ext.Close()
	if err := ext.SendHello(); err != nil {
		return resp, err
	}
	resp.WSHelloOK = true
	time.Sleep(50 * time.Millisecond)

	if err := fillSessionProbePair(resp, srv.BaseURL, req.SessionIDA, req.SessionIDB); err != nil {
		return resp, err
	}
	return resp, nil
}

func runWSDisconnectIsolation(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL}

	version := req.HelloVersion
	if version == "" {
		version = "1.0.0"
	}
	features := req.HelloFeatures
	if features == nil {
		features = []string{"browser-agent"}
	}

	extA, err := dialFakeExtension(srv.BaseURL, req.SessionIDA, version, features)
	if err != nil {
		return resp, err
	}
	extB, err := dialFakeExtension(srv.BaseURL, req.SessionIDB, version, features)
	if err != nil {
		extA.Close()
		return resp, err
	}
	if err := extA.SendHello(); err != nil {
		extA.Close()
		extB.Close()
		return resp, err
	}
	if err := extB.SendHello(); err != nil {
		extA.Close()
		extB.Close()
		return resp, err
	}

	// B auto-completes jobs; A does not.
	extB.AutoCompleteOK = true
	go extB.Loop()

	jobSeen := make(chan struct{}, 1)
	extA.OnJob = func(env wsEnvelope) {
		select {
		case jobSeen <- struct{}{}:
		default:
		}
	}
	go extA.Loop()
	time.Sleep(50 * time.Millisecond)

	timeoutMS := req.JobHTTPTimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 5000
	}
	jobType := req.JobHTTPType
	if jobType == "" {
		jobType = "eval"
	}
	params := req.JobHTTPParams
	if params == nil {
		params = map[string]any{"code": "hang"}
	}

	type postOut struct {
		status int
		body   []byte
		err    error
	}
	done := make(chan postOut, 1)
	go func() {
		st, _, body, perr := postJobs(srv.BaseURL, req.SessionIDA, jobType, params, timeoutMS)
		done <- postOut{st, body, perr}
	}()

	select {
	case <-jobSeen:
	case <-time.After(2 * time.Second):
	}

	extA.Close()
	resp.WSDisconnected = true

	select {
	case out := <-done:
		resp.StatusCode = out.status
		resp.Body = out.body
		resp.BodyString = string(out.body)
		if out.err != nil {
			return resp, out.err
		}
		parseJobHTTPResult(resp, out.body)
	case <-time.After(6 * time.Second):
		return resp, fmt.Errorf("POST /v1/jobs on A did not return after disconnect")
	}

	if err := fillSessionProbePair(resp, srv.BaseURL, req.SessionIDA, req.SessionIDB); err != nil {
		return resp, err
	}

	// B should still accept jobs after A disconnect.
	st, _, bodyB, perr := postJobs(srv.BaseURL, req.SessionIDB, "eval", map[string]any{"code": "2+2"}, 3000)
	if perr != nil {
		return resp, perr
	}
	resp.HTTPJobOKOnB = false
	parseJobHTTPResultInto(&resp.HTTPJobOKOnB, &resp.HTTPJobErrorOnB, bodyB)
	_ = st
	return resp, nil
}

func runWSJobRouting(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL}

	version := req.HelloVersion
	if version == "" {
		version = "1.0.0"
	}
	features := req.HelloFeatures
	if features == nil {
		features = []string{"browser-agent"}
	}

	extA, err := dialFakeExtension(srv.BaseURL, req.SessionIDA, version, features)
	if err != nil {
		return resp, err
	}
	defer extA.Close()
	extB, err := dialFakeExtension(srv.BaseURL, req.SessionIDB, version, features)
	if err != nil {
		return resp, err
	}
	defer extB.Close()

	if err := extA.SendHello(); err != nil {
		return resp, err
	}
	if err := extB.SendHello(); err != nil {
		return resp, err
	}
	resp.WSHelloOK = true
	time.Sleep(40 * time.Millisecond)

	jobChA := make(chan wsEnvelope, 1)
	jobChB := make(chan wsEnvelope, 1)
	extA.OnJob = func(env wsEnvelope) {
		select {
		case jobChA <- env:
		default:
		}
	}
	extB.OnJob = func(env wsEnvelope) {
		select {
		case jobChB <- env:
		default:
		}
	}
	go extA.Loop()
	go extB.Loop()

	target := req.JobTargetSessionID
	if target == "" {
		target = req.SessionIDA
	}

	go func() {
		_, _, _, _ = postJobs(srv.BaseURL, target, "eval", map[string]any{"code": "1"}, 5000)
	}()

	select {
	case env := <-jobChA:
		resp.WSJobReceivedOnA = true
		resp.WSJobType = envelopeJobType(env)
		b, _ := json.Marshal(env)
		resp.WSJobPayloadRaw = string(b)
	case <-time.After(3 * time.Second):
	}

	// Brief window: B must not receive the job meant for A.
	select {
	case env := <-jobChB:
		resp.WSJobReceivedOnB = true
		resp.WSJobType = envelopeJobType(env)
	case <-time.After(300 * time.Millisecond):
	}

	return resp, nil
}

// --- registry httptest harness ---

type registryHTTPServer struct {
	BaseURL  string
	registry *browseragent.SessionRegistry
	server   *httptest.Server
}

func startRegistryHTTPServer(t *testing.T, req *Request) (*registryHTTPServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by root Setup")
	}
	if req.SessionIDA == "" || req.SessionIDB == "" {
		t.Fatal("SessionIDA and SessionIDB must be set by root Setup")
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

	reg := browseragent.NewSessionRegistry(req.BaseDir, addr)
	ids := []string{req.SessionIDA, req.SessionIDB}
	for _, extra := range req.PreCreateSessionIDs {
		if extra != req.SessionIDA && extra != req.SessionIDB {
			ids = append(ids, extra)
		}
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		if _, err := reg.Create(id); err != nil {
			return nil, nil, fmt.Errorf("registry pre-create %q: %w", id, err)
		}
	}
	if req.SecondPreCreateID != "" && !seen[req.SecondPreCreateID] {
		if _, err := reg.Create(req.SecondPreCreateID); err != nil {
			return nil, nil, fmt.Errorf("registry pre-create %q: %w", req.SecondPreCreateID, err)
		}
	}

	handler := browseragent.NewRegistryControlHandler(reg)
	ts := httptest.NewServer(handler)

	out := &registryHTTPServer{
		BaseURL:  ts.URL,
		registry: reg,
		server:   ts,
	}
	cleanup := func() { ts.Close() }
	return out, cleanup, nil
}

func probeWSDial(baseURL, sessionID string) (int, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return 0, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
	if sessionID != "" {
		q := u.Query()
		q.Set("session", sessionID)
		u.RawQuery = q.Encode()
	}
	dialer := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
	conn, resp, err := dialer.Dial(u.String(), nil)
	if conn != nil {
		_ = conn.Close()
	}
	if resp != nil {
		return resp.StatusCode, err
	}
	return 0, err
}

func postJobs(baseURL, sessionID, jobType string, params map[string]any, timeoutMS int64) (int, string, []byte, error) {
	payload := map[string]any{
		"session_id": sessionID,
		"type":       jobType,
		"params":     params,
		"timeout_ms": timeoutMS,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, "", nil, err
	}
	clientTimeout := time.Duration(timeoutMS)*time.Millisecond + 2*time.Second
	if clientTimeout < 3*time.Second {
		clientTimeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/jobs", bytes.NewReader(b))
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

func fillSessionProbePair(resp *Response, baseURL, sessionA, sessionB string) error {
	uA := baseURL + "/v1/session?session=" + url.QueryEscape(sessionA)
	stA, _, bodyA, err := doGET(uA)
	if err != nil {
		return err
	}
	resp.SessionAProbeURL = uA
	resp.StatusCode = stA
	parseSessionFields(bodyA, &resp.ExtensionConnectedA, &resp.SupportsA)

	uB := baseURL + "/v1/session?session=" + url.QueryEscape(sessionB)
	stB, _, bodyB, err := doGET(uB)
	if err != nil {
		return err
	}
	resp.SessionBProbeURL = uB
	_ = stB
	parseSessionFields(bodyB, &resp.ExtensionConnectedB, &resp.SupportsB)
	return nil
}

func parseSessionFields(body []byte, connected, supports *bool) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	if ext, ok := raw["extension"].(map[string]any); ok {
		*connected, _ = ext["connected"].(bool)
		if v, ok := ext["supports_browser_agent"].(bool); ok {
			*supports = v
		}
	}
}

func parseJobHTTPResult(resp *Response, body []byte) {
	parseJobHTTPResultInto(&resp.HTTPJobOK, &resp.HTTPJobError, body)
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.Raw = raw
}

func parseJobHTTPResultInto(ok *bool, errMsg *string, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	src := raw
	if nested, ok := raw["result"].(map[string]any); ok {
		src = nested
	}
	if v, exists := src["ok"].(bool); exists {
		*ok = v
	}
	if e, ok := src["error"].(string); ok {
		*errMsg = e
	}
	if *errMsg == "" {
		if e, ok := raw["error"].(string); ok {
			*errMsg = e
		}
	}
}

// --- fake extension WS client (browser-agent harness pattern) ---

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

func envelopeJobType(env wsEnvelope) string {
	if env.Payload == nil {
		return ""
	}
	if t, ok := env.Payload["type"].(string); ok {
		return t
	}
	return ""
}

var (
	_ = sync.Mutex{}
	_ = io.Discard
)
```