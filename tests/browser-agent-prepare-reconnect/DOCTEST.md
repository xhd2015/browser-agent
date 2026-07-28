# browser-agent prepare_reconnect + prepare-upgrade (Phase 2)

Classic TDD for **Phase 2 of 3** daemon upgrade path helpers:

1. **WS envelope** `type=prepare_reconnect` with payload `reason`, `delay_ms` (default
   **1000**), and optional retry hints.
2. **Control-plane broadcast** of `prepare_reconnect` to every session with a live
   extension WebSocket; returns notified session ids.
3. **`POST /v1/admin/prepare-upgrade`** (documented admin path) triggers that broadcast
   and returns notified ids in the HTTP body.
4. **Chrome + Firefox** extension background sources handle `prepare_reconnect`:
   schedule WS close after `delay_ms`, then aggressive reconnect retries.

**Out of scope (Phase 3):** EnsureDaemon full orchestration and the 30s reattach wait
loop.

| Surface | What is under test |
|---------|-------------------|
| Message shape | Pure `BuildPrepareReconnectPayload` — type contract + defaults |
| Broadcast | Package broadcast → notified ids + WS delivery |
| prepare-upgrade HTTP | `POST /v1/admin/prepare-upgrade` → notified + WS envelope |
| Ext source markers | Chrome/Firefox `background.js` handle + delay + reconnect |

**No real Chrome/Firefox UI.** Fake extension WS via gorilla dialer for broadcast/HTTP
leaves. Ext-source leaves are static filesystem marker probes.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Control Server** hosts multi-session HTTP + per-session extension WebSockets
(`GET /v1/ws?session=<id>`). Each **Session** may have zero or one live **Extension WS**.

Before a daemon upgrade (Phase 3 orchestration), the control plane **prepares reconnect**:
it tells every connected extension to close its socket after a short delay so the
replacement daemon can accept fresh hellos.

**Prepare Reconnect Envelope** is a versioned WS message:

```text
{ v: 1, type: "prepare_reconnect", id: <msg-id>,
  payload: { reason?, delay_ms (default 1000), retry_base_ms?, retry_max_ms? } }
```

**Broadcast Prepare Reconnect** walks live sessions, writes the envelope to each
connected WS, and collects **notified session ids** (successful writes only).
Sessions without a WS are skipped.

**Admin Prepare Upgrade** (`POST /v1/admin/prepare-upgrade`) is the HTTP entry point
used by EnsureDaemon (Phase 3). Optional JSON body supplies `reason` / `delay_ms` /
retry hints. Response includes `notified` session id list.

**Extension Background** (Chrome shell + Firefox shell) on `prepare_reconnect`:
read `delay_ms`, **schedule close** of the session WS after that delay, then run
**aggressive reconnect** retries (lower backoff / force reconnect) so the extension
reattaches quickly after the new daemon listens.

**Test Client** builds pure payloads, starts `httptest` registry handlers with fake
extensions, POSTs admin upgrade, and probes extension source markers under ModuleRoot.

```text
# pure shape
BuildPrepareReconnectPayload(opts) -> payload with delay_ms default 1000

# broadcast
sessions with WS <- prepare_reconnect envelope
  -> notified = [session ids write-ok]

# HTTP
POST /v1/admin/prepare-upgrade {reason?, delay_ms?}
  -> 200 { ok, notified: [...] }
  -> each connected WS receives type=prepare_reconnect

# extension
on message type prepare_reconnect
  -> setTimeout(delay_ms) -> ws.close()
  -> aggressive reconnect (retry_base_ms / force reconnect)
```

## Decision Tree

```
browser-agent-prepare-reconnect
├── message-shape/                           [pure payload builder]
│   ├── default-delay-ms/                        DelayMS<=0 → delay_ms=1000
│   ├── custom-delay-and-reason/                 explicit delay_ms + reason
│   └── with-retry-hints/                        retry_base_ms + retry_max_ms
├── broadcast/                               [package broadcast + WS delivery]
│   ├── single-connected/                        one WS → notified [id] + envelope
│   ├── multi-connected/                         two WS → both notified + both receive
│   ├── only-connected-notified/                 connected+orphan → only connected id
│   └── zero-connected/                          no WS → notified []
├── prepare-upgrade-http/                    [POST /v1/admin/prepare-upgrade]
│   ├── post-notifies-and-delivers/              200 + notified + WS type
│   ├── post-empty-notified/                     no WS → 200 + notified []
│   └── get-method-not-allowed/                  GET → 405
└── ext-source/                              [Chrome + Firefox background markers]
    ├── chrome-background/                       prepare_reconnect + delay + reconnect
    └── firefox-background/                      same markers in Firefox shell
```

### Parameter significance (high → low)

1. **Surface** — pure message vs broadcast vs admin HTTP vs extension source.
2. **Within message-shape** — default delay vs explicit fields vs optional retry hints.
3. **Within broadcast** — connectivity topology (0 / 1 / multi / mix).
4. **Within prepare-upgrade-http** — success delivery vs empty vs method guard.
5. **Within ext-source** — browser product (Chrome shell vs Firefox shell).

## Test Index

| Leaf | Scenario |
|------|----------|
| `message-shape/default-delay-ms` | `DelayMS<=0` → payload `delay_ms=1000`; no crash |
| `message-shape/custom-delay-and-reason` | Custom `delay_ms` + non-empty `reason` in payload |
| `message-shape/with-retry-hints` | `retry_base_ms` + `retry_max_ms` present when set |
| `broadcast/single-connected` | One hello WS → notified contains that id; WS sees `prepare_reconnect` |
| `broadcast/multi-connected` | Two WS → both ids notified; both receive envelope |
| `broadcast/only-connected-notified` | Connected A + disconnected B → only A notified |
| `broadcast/zero-connected` | Sessions exist, no WS → empty notified, no error |
| `prepare-upgrade-http/post-notifies-and-delivers` | POST admin → 200, notified, WS type=`prepare_reconnect` + delay_ms |
| `prepare-upgrade-http/post-empty-notified` | POST with no WS → 200, `notified` empty |
| `prepare-upgrade-http/get-method-not-allowed` | GET admin path → 405 |
| `ext-source/chrome-background` | Chrome `background.js` markers: prepare_reconnect, delay_ms, close+reconnect |
| `ext-source/firefox-background` | Firefox `background.js` same markers |

**Leaf count: 12**

## How to Run

```sh
doctest vet ./tests/browser-agent-prepare-reconnect
doctest test ./tests/browser-agent-prepare-reconnect   # expect RED (Classic TDD)
```

Tree is **RED** until implementer lands Phase 2 APIs + extension handlers.

### Implementer contract (authoritative for GREEN)

```text
// DefaultPrepareReconnectDelayMS is used when DelayMS <= 0.
const DefaultPrepareReconnectDelayMS = 1000

// PrepareReconnectOptions configures prepare_reconnect payload / broadcast.
type PrepareReconnectOptions struct {
    Reason      string // optional; e.g. "daemon-upgrade"
    DelayMS     int    // <=0 → DefaultPrepareReconnectDelayMS (1000)
    RetryBaseMS int    // optional; >0 → payload["retry_base_ms"]
    RetryMaxMS  int    // optional; >0 → payload["retry_max_ms"]
}

// BuildPrepareReconnectPayload returns the WS payload map for type=prepare_reconnect.
// Always includes delay_ms (int). Includes reason when non-empty.
// Includes retry_base_ms / retry_max_ms only when > 0.
func BuildPrepareReconnectPayload(opts PrepareReconnectOptions) map[string]any

// BroadcastPrepareReconnect writes type=prepare_reconnect (v=1) to every session
// that currently has a live extension WebSocket writer.
// Returns sorted session ids for which the write succeeded.
// Sessions without WS are skipped (not an error).
// delay_ms / reason / retry hints follow BuildPrepareReconnectPayload rules.
func BroadcastPrepareReconnect(r *SessionRegistry, opts PrepareReconnectOptions) []string

// HTTP
// POST /v1/admin/prepare-upgrade
//   optional JSON body: { reason?, delay_ms?, retry_base_ms?, retry_max_ms? }
//   200 application/json:
//     { "ok": true, "notified": ["sess-…", ...] }  // sorted ids, may be empty
//   only POST allowed; other methods → 405
// Implementation: parse body → BroadcastPrepareReconnect → encode notified.

// WS envelope delivered to extensions:
// {
//   "v": 1,
//   "type": "prepare_reconnect",
//   "id": "<non-empty>",
//   "payload": { "delay_ms": <int>, "reason"?: string, "retry_base_ms"?: int, "retry_max_ms"?: int }
// }

// Extension sources (Chrome-Ext-Browser-Agent/public/background.js and
// Firefox-Ext-Browser-Agent/public/background.js — build/src fallbacks OK):
// - handle incoming type "prepare_reconnect"
// - read payload.delay_ms (or delayMs)
// - schedule close of the session WebSocket after delay_ms
// - aggressive reconnect after close (reset backoff / force reconnect / lower base)
```

```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
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
	ModeMessageShape       = "message-shape"
	ModeBroadcast          = "broadcast"
	ModePrepareUpgradeHTTP = "prepare-upgrade-http"
	ModeExtSource          = "ext-source"
)

// MessageShapeCase — pure payload builder scenarios.
const (
	MessageShapeDefaultDelay   = "default-delay-ms"
	MessageShapeCustomDelay    = "custom-delay-and-reason"
	MessageShapeWithRetryHints = "with-retry-hints"
)

// BroadcastCase — connectivity topology scenarios.
const (
	BroadcastSingleConnected       = "single-connected"
	BroadcastMultiConnected        = "multi-connected"
	BroadcastOnlyConnectedNotified = "only-connected-notified"
	BroadcastZeroConnected         = "zero-connected"
)

// PrepareUpgradeHTTPCase — admin HTTP scenarios.
const (
	PrepareUpgradePostNotifies = "post-notifies-and-delivers"
	PrepareUpgradePostEmpty    = "post-empty-notified"
	PrepareUpgradeGetNotAllowed = "get-method-not-allowed"
)

// ExtSourceTarget — browser product background under ModuleRoot.
const (
	ExtSrcChromeBackground  = "chrome-background"
	ExtSrcFirefoxBackground = "firefox-background"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	// message-shape
	MessageShapeCase string
	// Fields fed into BuildPrepareReconnectPayload (via package API).
	Reason      string
	DelayMS     int
	RetryBaseMS int
	RetryMaxMS  int

	// broadcast
	BroadcastCase       string
	SessionIDs          []string // sessions to Create in registry
	ConnectSessionIDs   []string // subset that dial WS + hello
	BroadcastReason     string
	BroadcastDelayMS    int

	// prepare-upgrade-http
	PrepareUpgradeCase string
	HTTPMethod         string // default POST; GET for method-not-allowed
	ConnectForHTTP     bool
	HTTPBodyReason     string
	HTTPBodyDelayMS    int
	// When true, include delay_ms in POST body even if 0 is meaningful — leaves set explicitly.
	HTTPBodyIncludeDelay bool

	// ext-source
	ExtSourceTarget string

	// Shared hello identity for fake extensions
	HelloVersion  string
	HelloFeatures []string

	// Timeouts
	WSReceiveTimeout time.Duration
}

// Response holds outcomes for all modes.
type Response struct {
	// message-shape
	Payload       map[string]any
	PayloadJSON   string
	DelayMSValue  int
	ReasonValue   string
	HasRetryBase  bool
	HasRetryMax   bool
	RetryBaseValue int
	RetryMaxValue  int

	// broadcast
	NotifiedIDs          []string
	WSReceivedBySession  map[string]bool
	WSTypeBySession      map[string]string
	WSDelayMSBySession   map[string]int
	WSPayloadRawBySession map[string]string
	BroadcastErrText     string

	// prepare-upgrade-http
	StatusCode      int
	ContentType     string
	Body            []byte
	BodyString      string
	HTTPNotifiedIDs []string
	HTTPOK          bool
	HTTPWSType      string
	HTTPWSDelayMS   int
	HTTPWSReceived  bool
	BaseURL         string
	ProbeURL        string

	// ext-source
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	ErrText      string

	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}
	switch req.Mode {
	case ModeMessageShape:
		return runMessageShape(t, req)
	case ModeBroadcast:
		return runBroadcast(t, req)
	case ModePrepareUpgradeHTTP:
		return runPrepareUpgradeHTTP(t, req)
	case ModeExtSource:
		return runExtSource(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runMessageShape(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.MessageShapeCase == "" {
		t.Fatal("MessageShapeCase must be set by leaf Setup")
	}
	opts := browseragent.PrepareReconnectOptions{
		Reason:      req.Reason,
		DelayMS:     req.DelayMS,
		RetryBaseMS: req.RetryBaseMS,
		RetryMaxMS:  req.RetryMaxMS,
	}
	payload := browseragent.BuildPrepareReconnectPayload(opts)
	resp := &Response{
		ExitCode: 0,
		Payload:  payload,
	}
	if payload != nil {
		b, err := json.Marshal(payload)
		if err == nil {
			resp.PayloadJSON = string(b)
		}
		resp.DelayMSValue = payloadInt(payload, "delay_ms")
		if s, ok := payload["reason"].(string); ok {
			resp.ReasonValue = s
		}
		if v, ok := payloadNum(payload, "retry_base_ms"); ok {
			resp.HasRetryBase = true
			resp.RetryBaseValue = v
		}
		if v, ok := payloadNum(payload, "retry_max_ms"); ok {
			resp.HasRetryMax = true
			resp.RetryMaxValue = v
		}
	}
	return resp, nil
}

func runBroadcast(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BroadcastCase == "" {
		t.Fatal("BroadcastCase must be set by leaf Setup")
	}
	srv, cleanup, err := startRegistryHTTPServer(t, req, req.SessionIDs)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		ExitCode:              0,
		BaseURL:               srv.BaseURL,
		WSReceivedBySession:   map[string]bool{},
		WSTypeBySession:       map[string]string{},
		WSDelayMSBySession:    map[string]int{},
		WSPayloadRawBySession: map[string]string{},
	}

	recvTimeout := req.WSReceiveTimeout
	if recvTimeout <= 0 {
		recvTimeout = 2 * time.Second
	}

	// Channels per connected session.
	type slot struct {
		id  string
		ext *fakeExtension
		ch  chan wsEnvelope
	}
	var slots []slot
	for _, id := range req.ConnectSessionIDs {
		ext, err := dialFakeExtension(srv.BaseURL, id, req.HelloVersion, req.HelloFeatures)
		if err != nil {
			return resp, fmt.Errorf("dial fake ext %s: %w", id, err)
		}
		defer ext.Close()
		if err := ext.SendHello(); err != nil {
			return resp, fmt.Errorf("hello %s: %w", id, err)
		}
		ch := make(chan wsEnvelope, 2)
		sid := id
		ext.OnPrepareReconnect = func(env wsEnvelope) {
			select {
			case ch <- env:
			default:
			}
		}
		// Also capture any envelope via generic path if implementer uses different type spelling — Loop handles prepare_reconnect.
		go ext.Loop()
		slots = append(slots, slot{id: sid, ext: ext, ch: ch})
	}
	if len(slots) > 0 {
		time.Sleep(50 * time.Millisecond)
	}

	opts := browseragent.PrepareReconnectOptions{
		Reason:  req.BroadcastReason,
		DelayMS: req.BroadcastDelayMS,
	}
	notified := browseragent.BroadcastPrepareReconnect(srv.registry, opts)
	resp.NotifiedIDs = append([]string(nil), notified...)

	// Collect WS deliveries.
	deadline := time.Now().Add(recvTimeout)
	for _, s := range slots {
		wait := time.Until(deadline)
		if wait < 0 {
			wait = 0
		}
		select {
		case env := <-s.ch:
			resp.WSReceivedBySession[s.id] = true
			resp.WSTypeBySession[s.id] = env.Type
			if env.Payload != nil {
				resp.WSDelayMSBySession[s.id] = payloadInt(env.Payload, "delay_ms")
			}
			b, _ := json.Marshal(env)
			resp.WSPayloadRawBySession[s.id] = string(b)
		case <-time.After(wait):
			resp.WSReceivedBySession[s.id] = false
		}
	}
	return resp, nil
}

func runPrepareUpgradeHTTP(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.PrepareUpgradeCase == "" {
		t.Fatal("PrepareUpgradeCase must be set by leaf Setup")
	}
	sessionIDs := req.SessionIDs
	if len(sessionIDs) == 0 {
		sessionIDs = []string{"sess-admin-a"}
	}
	srv, cleanup, err := startRegistryHTTPServer(t, req, sessionIDs)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		ExitCode: 0,
		BaseURL:  srv.BaseURL,
		ProbeURL: srv.BaseURL + "/v1/admin/prepare-upgrade",
	}

	method := req.HTTPMethod
	if method == "" {
		method = http.MethodPost
	}

	// GET method-not-allowed: no need for WS.
	if method != http.MethodPost {
		status, ct, body, err := doHTTP(method, resp.ProbeURL, nil)
		if err != nil {
			return resp, err
		}
		resp.StatusCode = status
		resp.ContentType = ct
		resp.Body = body
		resp.BodyString = string(body)
		return resp, nil
	}

	var ext *fakeExtension
	var ch chan wsEnvelope
	if req.ConnectForHTTP {
		id := sessionIDs[0]
		ext, err = dialFakeExtension(srv.BaseURL, id, req.HelloVersion, req.HelloFeatures)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		if err := ext.SendHello(); err != nil {
			return resp, err
		}
		ch = make(chan wsEnvelope, 2)
		ext.OnPrepareReconnect = func(env wsEnvelope) {
			select {
			case ch <- env:
			default:
			}
		}
		go ext.Loop()
		time.Sleep(50 * time.Millisecond)
	}

	bodyMap := map[string]any{}
	if req.HTTPBodyReason != "" {
		bodyMap["reason"] = req.HTTPBodyReason
	}
	if req.HTTPBodyIncludeDelay || req.HTTPBodyDelayMS > 0 {
		bodyMap["delay_ms"] = req.HTTPBodyDelayMS
	}
	var body any
	if len(bodyMap) > 0 {
		body = bodyMap
	}
	status, ct, raw, err := doHTTP(http.MethodPost, resp.ProbeURL, body)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = raw
	resp.BodyString = string(raw)
	parsePrepareUpgradeJSON(resp, raw)

	if ch != nil {
		timeout := req.WSReceiveTimeout
		if timeout <= 0 {
			timeout = 2 * time.Second
		}
		select {
		case env := <-ch:
			resp.HTTPWSReceived = true
			resp.HTTPWSType = env.Type
			if env.Payload != nil {
				resp.HTTPWSDelayMS = payloadInt(env.Payload, "delay_ms")
			}
		case <-time.After(timeout):
			resp.HTTPWSReceived = false
		}
	}
	return resp, nil
}

func runExtSource(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ExtSourceTarget == "" {
		t.Fatal("ExtSourceTarget must be set by leaf Setup")
	}
	root := req.ModuleRoot
	resp := &Response{ExitCode: 0}

	var candidates []string
	switch req.ExtSourceTarget {
	case ExtSrcChromeBackground:
		candidates = chromeBackgroundCandidates(root)
	case ExtSrcFirefoxBackground:
		candidates = firefoxBackgroundCandidates(root)
	default:
		return nil, fmt.Errorf("unknown ExtSourceTarget %q", req.ExtSourceTarget)
	}

	path, data, ok := firstExistingFile(candidates)
	resp.FileExists = ok
	if ok {
		resp.FoundPaths = []string{path}
		resp.CombinedText = string(data)
	} else {
		resp.ErrText = "background.js not found for " + req.ExtSourceTarget
	}
	return resp, nil
}

// --- registry httptest harness ---

type registryHTTPServer struct {
	BaseURL  string
	registry *browseragent.SessionRegistry
	server   *httptest.Server
}

func startRegistryHTTPServer(t *testing.T, req *Request, sessionIDs []string) (*registryHTTPServer, func(), error) {
	t.Helper()
	baseDir := req.BaseDir
	if baseDir == "" {
		baseDir = t.TempDir()
		req.BaseDir = baseDir
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

	reg := browseragent.NewSessionRegistry(baseDir, addr)
	for _, id := range sessionIDs {
		if id == "" {
			continue
		}
		if _, err := reg.Create(id); err != nil {
			return nil, nil, fmt.Errorf("registry Create %q: %w", id, err)
		}
	}

	h := browseragent.NewRegistryControlHandler(reg)
	srv := httptest.NewServer(h)
	return &registryHTTPServer{
		BaseURL:  srv.URL,
		registry: reg,
		server:   srv,
	}, srv.Close, nil
}

// --- fake extension WS client ---

type wsEnvelope struct {
	V       int            `json:"v"`
	Type    string         `json:"type"`
	ID      string         `json:"id"`
	Payload map[string]any `json:"payload"`
}

type fakeExtension struct {
	conn               *websocket.Conn
	version            string
	features           []string
	OnPrepareReconnect func(wsEnvelope)
	OnJob              func(wsEnvelope)
	mu                 sync.Mutex
	closed             bool
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
		case "prepare_reconnect":
			if f.OnPrepareReconnect != nil {
				f.OnPrepareReconnect(env)
			}
		case "job":
			if f.OnJob != nil {
				f.OnJob(env)
			}
		case "ping":
			_ = f.conn.WriteJSON(wsEnvelope{V: 1, Type: "pong", ID: env.ID})
		case "status":
			// hello ack — ignore
		}
	}
}

func (f *fakeExtension) Close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	_ = f.conn.Close()
}

// --- HTTP helpers ---

func doHTTP(method, rawURL string, body any) (status int, contentType string, raw []byte, err error) {
	var rdr io.Reader
	if body != nil {
		b, mErr := json.Marshal(body)
		if mErr != nil {
			return 0, "", nil, mErr
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, rawURL, rdr)
	if err != nil {
		return 0, "", nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer res.Body.Close()
	raw, err = io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, res.Header.Get("Content-Type"), nil, err
	}
	return res.StatusCode, res.Header.Get("Content-Type"), raw, nil
}

func parsePrepareUpgradeJSON(resp *Response, raw []byte) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	if ok, _ := m["ok"].(bool); ok {
		resp.HTTPOK = true
	}
	switch n := m["notified"].(type) {
	case []any:
		for _, item := range n {
			if s, ok := item.(string); ok && s != "" {
				resp.HTTPNotifiedIDs = append(resp.HTTPNotifiedIDs, s)
			}
		}
	case []string:
		resp.HTTPNotifiedIDs = append(resp.HTTPNotifiedIDs, n...)
	}
	sort.Strings(resp.HTTPNotifiedIDs)
}

// --- payload helpers ---

func payloadInt(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	default:
		return 0
	}
}

func payloadNum(m map[string]any, key string) (int, bool) {
	if m == nil {
		return 0, false
	}
	if _, ok := m[key]; !ok {
		return 0, false
	}
	return payloadInt(m, key), true
}

// --- ext source paths ---

func chromeBackgroundCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "background.js"),
		filepath.Join(root, "browseragent", "embedded", "extension", "background.js"),
	}
}

func firefoxBackgroundCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "public", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "src", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "build", "background.js"),
		filepath.Join(root, "browseragent", "embedded", "extension-firefox", "background.js"),
	}
}

func firstExistingFile(paths []string) (string, []byte, bool) {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return p, data, true
		}
	}
	return "", nil, false
}

// silence unused import if helpers evolve
var (
	_ = strings.Contains
	_ = sort.Strings
)
```
