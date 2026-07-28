# browser-agent Phase 1 — HTTP poll transport for extension (server)

Classic TDD for **Phase 1** of extension attach over plain HTTP when WebSocket is
unavailable (Firefox path falls back in Phase 2). The daemon gains three control
routes that share the same session registry / `markHello` / job queue as WS:

| Route | Role |
|-------|------|
| `POST /v1/ext/hello` | Attach: mark extension connected (no WS required) |
| `POST /v1/ext/poll` | Long-poll: lease queued jobs + drain session events |
| `POST /v1/ext/result` | Complete a job (same as WS `type=result`) |

**No real browser.** Integration via `httptest.Server` + `SessionRegistry` +
`NewRegistryControlHandler` + registry `Create`. Pure `NormalizeExtPollWaitMS`
lives in the nested tree `wait-ms/` (own `DOCTEST.md`) so HTTP leaves still
**compile** while the helper is missing. Jobs are enqueued with `POST /v1/jobs`
(after hello so `fastFailNoExtension` does not reject); without a WS, push is
skipped and the job stays **queued** until poll leases it.

**Out of scope (Phase 2):** Firefox extension JS poll loop.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Control Server** hosts multi-session HTTP (`NewRegistryControlHandler`). Each
**Session** lives in **SessionRegistry** with a **JobQueue** and optional
**event queue** for poll-mode control messages.

**HTTP Extension Transport** (this phase) is an alternative to **WebSocket
Extension** attach:

1. **Hello** — extension POSTs `session_id` (+ optional `version`, `features`,
   `bundle_md5`, `browser_product`). Server calls the same attach path as WS
   hello (`markHello` / phase `extension_connected`). No socket is required.
2. **Poll** — extension POSTs `session_id` and optional `wait_ms`. Server
   **leases** (dequeue → running) any queued jobs for that session and returns
   pending **events** (e.g. `prepare_reconnect`). If nothing is ready and
   `wait_ms > 0`, the handler **blocks** until a job/event appears or the wait
   elapses (then empty `jobs` / `events`).
3. **Result** — extension POSTs `session_id`, `job_id`, `ok`, optional `data` /
   `error`. Server completes the job like the WS result handler so
   `POST /v1/jobs` waiters unblock.

**WaitMS helper** normalizes poll wait: default **25000** ms when unset/`<=0`,
hard **cap 30000** ms.

**Test Client** creates a registry session, starts httptest, POSTs the three
routes, enqueues work via `POST /v1/jobs` in a background goroutine when
needed, and probes `GET /v1/session?session=` for connected state.

```text
registry.Create(session_id)
POST /v1/ext/hello {session_id, version?, features?}
  -> 200 {ok:true, phase:"extension_connected"}
  -> GET /v1/session shows extension.connected=true

POST /v1/jobs (background) enqueues job (no WS -> stays Queued)
POST /v1/ext/poll {session_id, wait_ms}
  -> 200 {jobs:[{id,type,...}], events:[]}
POST /v1/ext/result {session_id, job_id, ok:true, data?}
  -> 200 {ok:true}
  -> background /v1/jobs returns ok=true

BroadcastPrepareReconnect / prepare-upgrade
  -> poll events include type prepare_reconnect for HTTP-connected sessions
```

## Decision Tree

```
browser-agent-ext-http-poll
├── hello/                                   [POST /v1/ext/hello]
│   ├── connects/                                known session → 200 + phase + connected
│   ├── unknown-404/                             unknown session_id → 404
│   └── missing-session-id-400/                  empty/missing session_id → 400
├── poll/                                    [POST /v1/ext/poll]
│   ├── returns-queued-job/                      hello + enqueue → jobs non-empty
│   ├── empty-on-timeout/                        short wait_ms, no work → empty arrays
│   ├── prepare-reconnect-event/                 hello + broadcast → events type
│   └── unknown-404/                             unknown session_id → 404
├── result/                                  [POST /v1/ext/result]
│   ├── completes-ok/                            poll job + result → /v1/jobs ok
│   └── unknown-404/                             unknown session_id → 404
├── flow/                                    [end-to-end happy path]
│   └── hello-poll-result/                       hello → enqueue → poll → result → done
└── wait-ms/                                 [nested DOCTEST — pure NormalizeExtPollWaitMS]
    ├── default/                                 <=0 → 25000
    ├── within-cap/                              5000 → 5000
    └── over-cap/                                99999 → 30000
```

### Parameter significance (high → low)

1. **Surface** — hello vs poll vs result vs full flow (HTTP tree); wait_ms nested pure tree.
2. **Session validity** — known vs unknown vs missing `session_id`.
3. **Work available** — queued job vs empty timeout vs prepare_reconnect event.
4. **Outcome** — HTTP status, body fields, session connected, job completion.

## Test Index

| Leaf | Scenario |
|------|----------|
| `hello/connects` | Known session → **200**, `ok`, `phase=extension_connected`, GET session connected |
| `hello/unknown-404` | Unknown `session_id` → **404** |
| `hello/missing-session-id-400` | Missing/empty `session_id` → **400** |
| `poll/returns-queued-job` | After hello + background job enqueue → poll returns job with matching type |
| `poll/empty-on-timeout` | Hello only; `wait_ms=100` → **200** with empty `jobs` and `events` |
| `poll/prepare-reconnect-event` | Hello + `BroadcastPrepareReconnect` → poll `events` include `prepare_reconnect` |
| `poll/unknown-404` | Unknown session → **404** |
| `result/completes-ok` | Hello + enqueue + poll + result `ok:true` → `/v1/jobs` waiter succeeds |
| `result/unknown-404` | Unknown session on result → **404** |
| `flow/hello-poll-result` | Full round-trip: hello connects, poll leases job, result completes |
| `wait-ms/*` (nested) | See `wait-ms/DOCTEST.md` — default / within-cap / over-cap |

**Leaf count: 10** (this root) + **3** (nested `wait-ms/`)

## How to Run

```sh
doctest vet ./tests/browser-agent-ext-http-poll
doctest test ./tests/browser-agent-ext-http-poll           # HTTP root: 10 leaves, expect RED
doctest test ./tests/browser-agent-ext-http-poll/wait-ms   # nested pure helper: build RED until symbol exists
# Note: `doctest test ./tests/browser-agent-ext-http-poll/...` also pulls wait-ms;
# until NormalizeExtPollWaitMS exists the combined suite fails to compile.
```

Tree is **RED** until implementer lands Phase 1 routes + wait_ms helper.

### Implementer contract (authoritative for GREEN)

```text
// Defaults for POST /v1/ext/poll wait_ms.
const DefaultExtPollWaitMS = 25000
const MaxExtPollWaitMS     = 30000

// NormalizeExtPollWaitMS returns DefaultExtPollWaitMS when waitMS <= 0,
// MaxExtPollWaitMS when waitMS > MaxExtPollWaitMS, otherwise waitMS.
func NormalizeExtPollWaitMS(waitMS int) int

// --- POST /v1/ext/hello ---
// Body JSON: {
//   "session_id": string,           // required
//   "version"?: string,
//   "features"?: []string,
//   "bundle_md5"?: string,          // also accept bundleMd5 / md5
//   "browser_product"?: string
// }
// 200 application/json: { "ok": true, "phase": "extension_connected" }
//   (+ optional "session" / session_id echo)
// Side effect: same as WS hello — markHello(version, features, bundle_md5);
//   optional telemetry browser_product. Session becomes extension.connected.
// 400 if session_id missing/empty.
// 404 if session unknown.
// Only POST; other methods → 405 (not required by this tree).

// --- POST /v1/ext/poll ---
// Body JSON: {
//   "session_id": string,           // required
//   "wait_ms"?: int                 // normalized via NormalizeExtPollWaitMS
// }
// 200 application/json: {
//   "jobs": [ {
//       "id"|"job_id": string,
//       "type": string,
//       "params"?: object,
//       "timeout_ms"?: number,
//       "tab_id"?: number,
//       "session_id"?: string
//   }, ... ],
//   "events": [ {
//       "type": string,             // e.g. "prepare_reconnect"
//       "payload"?: object,         // delay_ms, reason, ...
//       "id"?: string
//   }, ... ]
// }
// Behavior:
// - Lease/dequeue queued jobs for the session (status → running), return them.
// - Drain pending poll events for the session into events[].
// - If both empty and wait_ms > 0 after normalize, block up to wait_ms for
//   new work; on timeout return empty arrays (still 200).
// - HTTP-connected sessions (hello via this route, no WS) must receive
//   prepare_reconnect as poll events when BroadcastPrepareReconnect runs
//   (extend broadcast / session event queue). WS-only delivery is insufficient.
// 400 if session_id missing/empty (optional for tree; unknown covered).
// 404 if session unknown.
// Prefer not requiring a prior hello for 404/empty; job delivery leaves call hello first.
// Soft auto-connect on poll is allowed but not required by asserts.

// --- POST /v1/ext/result ---
// Body JSON: {
//   "session_id": string,           // required
//   "job_id": string,               // required (also accept "id")
//   "ok": bool,
//   "data"?: object,
//   "error"?: string
// }
// 200 application/json: { "ok": true }
// Side effect: queue.Complete(job_id, JobResult{...}) like handleWSResult.
// 400 if session_id or job_id missing (not asserted in this tree).
// 404 if session unknown.

// Routes registered on NewRegistryControlHandler mux:
//   POST /v1/ext/hello
//   POST /v1/ext/poll
//   POST /v1/ext/result
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
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface under test (HTTP routes only; pure wait-ms is nested).
const (
	ModeHello  = "hello"
	ModePoll   = "poll"
	ModeResult = "result"
	ModeFlow   = "flow"
)

// HelloCase — POST /v1/ext/hello scenarios.
const (
	HelloConnects            = "connects"
	HelloUnknown404          = "unknown-404"
	HelloMissingSessionID400 = "missing-session-id-400"
)

// PollCase — POST /v1/ext/poll scenarios.
const (
	PollReturnsQueuedJob       = "returns-queued-job"
	PollEmptyOnTimeout         = "empty-on-timeout"
	PollPrepareReconnectEvent  = "prepare-reconnect-event"
	PollUnknown404             = "unknown-404"
)

// ResultCase — POST /v1/ext/result scenarios.
const (
	ResultCompletesOK = "completes-ok"
	ResultUnknown404  = "unknown-404"
)

// FlowCase — end-to-end scenarios.
const (
	FlowHelloPollResult = "hello-poll-result"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	// Session under test (created unless OmitCreate).
	SessionID   string
	OmitCreate  bool
	// Extra sessions to Create (unused in v0.0.2 leaves; reserved).
	ExtraSessionIDs []string

	// hello
	HelloCase       string
	HelloVersion    string
	HelloFeatures   []string
	HelloBundleMD5  string
	HelloBrowser    string
	// When true, omit session_id from hello body (or send empty).
	HelloOmitSessionID bool
	HelloSessionID     string // override body session_id when set (incl. wrong id)

	// poll
	PollCase       string
	PollWaitMS     int
	PollOmitWaitMS bool
	// When true, omit session_id from poll body.
	PollOmitSessionID bool
	PollSessionID     string // override body session_id

	// Before poll: enqueue a job via background POST /v1/jobs.
	EnqueueJob       bool
	JobHTTPType      string
	JobHTTPParams    map[string]any
	JobHTTPTimeoutMS int64

	// Before poll: BroadcastPrepareReconnect on registry.
	BroadcastPrepareReconnect bool
	BroadcastReason           string
	BroadcastDelayMS          int

	// result
	ResultCase       string
	ResultOK         bool
	ResultData       map[string]any
	ResultError      string
	ResultOmitSessionID bool
	ResultSessionID  string
	// When set, use this job_id instead of polled id (error leaves).
	ResultJobIDOverride string

	// flow
	FlowCase string

	// Shared timeouts for client HTTP.
	HTTPClientTimeout time.Duration
	JobWaitTimeout    time.Duration
}

// Response holds outcomes for all modes.
type Response struct {
	// HTTP generic
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string
	Raw         map[string]any

	// hello / session probe
	HTTPOK               bool
	Phase                string
	ExtensionConnected   bool
	SessionProbeURL      string
	SessionProbeStatus   int

	// poll
	Jobs              []map[string]any
	Events            []map[string]any
	JobCount          int
	EventCount        int
	FirstJobID        string
	FirstJobType      string
	HasPrepareReconnect bool
	PollElapsedMS     int64

	// result
	ResultOKField bool

	// background /v1/jobs waiter (enqueue + complete flows)
	HTTPJobStatus int
	HTTPJobOK     bool
	HTTPJobError  string
	HTTPJobID     string
	HTTPJobBody   string

	// meta
	BaseURL    string
	ProbeURL   string
	ExitCode   int
	RunErrText string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeHello:
		return runHello(t, req)
	case ModePoll:
		return runPoll(t, req)
	case ModeResult:
		return runResult(t, req)
	case ModeFlow:
		return runFlow(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runHello(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.HelloCase == "" {
		t.Fatal("HelloCase must be set by leaf Setup")
	}
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{ExitCode: 0, BaseURL: srv.BaseURL, ProbeURL: srv.BaseURL + "/v1/ext/hello"}

	body := map[string]any{}
	if !req.HelloOmitSessionID {
		sid := req.HelloSessionID
		if sid == "" {
			sid = req.SessionID
		}
		body["session_id"] = sid
	}
	if req.HelloVersion != "" {
		body["version"] = req.HelloVersion
	}
	if req.HelloFeatures != nil {
		body["features"] = req.HelloFeatures
	}
	if req.HelloBundleMD5 != "" {
		body["bundle_md5"] = req.HelloBundleMD5
	}
	if req.HelloBrowser != "" {
		body["browser_product"] = req.HelloBrowser
	}

	status, ct, raw, err := doJSON(http.MethodPost, resp.ProbeURL, body, req.HTTPClientTimeout)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = raw
	resp.BodyString = string(raw)
	parseHelloJSON(resp, raw)

	// Probe session only when we expect a known created session.
	if req.HelloCase == HelloConnects && req.SessionID != "" && !req.OmitCreate {
		u := srv.BaseURL + "/v1/session?session=" + url.QueryEscape(req.SessionID)
		st, _, sbody, gerr := doJSON(http.MethodGet, u, nil, req.HTTPClientTimeout)
		if gerr != nil {
			return resp, gerr
		}
		resp.SessionProbeURL = u
		resp.SessionProbeStatus = st
		parseSessionConnected(resp, sbody)
	}
	return resp, nil
}

func runPoll(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.PollCase == "" {
		t.Fatal("PollCase must be set by leaf Setup")
	}
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{ExitCode: 0, BaseURL: srv.BaseURL, ProbeURL: srv.BaseURL + "/v1/ext/poll"}

	// Known-session leaves: hello first so fastFailNoExtension allows jobs.
	if !req.OmitCreate && req.SessionID != "" && req.PollCase != PollUnknown404 {
		if err := postExtHello(srv.BaseURL, req); err != nil {
			return resp, fmt.Errorf("pre-hello: %w", err)
		}
	}

	var jobDone chan jobWaitOutcome
	if req.EnqueueJob {
		jobDone = startBackgroundJob(srv.BaseURL, req)
		// Brief yield so enqueue lands before poll.
		time.Sleep(30 * time.Millisecond)
	}

	if req.BroadcastPrepareReconnect {
		_ = browseragent.BroadcastPrepareReconnect(srv.registry, browseragent.PrepareReconnectOptions{
			Reason:  req.BroadcastReason,
			DelayMS: req.BroadcastDelayMS,
		})
	}

	body := map[string]any{}
	if !req.PollOmitSessionID {
		sid := req.PollSessionID
		if sid == "" {
			sid = req.SessionID
		}
		body["session_id"] = sid
	}
	if !req.PollOmitWaitMS {
		body["wait_ms"] = req.PollWaitMS
	}

	clientTimeout := req.HTTPClientTimeout
	if clientTimeout <= 0 {
		// Poll may block up to wait_ms; add slack.
		wm := req.PollWaitMS
		if wm <= 0 {
			wm = 1000
		}
		clientTimeout = time.Duration(wm)*time.Millisecond + 3*time.Second
	}

	started := time.Now()
	status, ct, raw, err := doJSON(http.MethodPost, resp.ProbeURL, body, clientTimeout)
	resp.PollElapsedMS = time.Since(started).Milliseconds()
	if err != nil {
		return resp, err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = raw
	resp.BodyString = string(raw)
	parsePollJSON(resp, raw)

	if jobDone != nil {
		// Do not require result in pure poll-job leaf; cancel wait by abandoning.
		// Drain with short timeout so goroutine does not leak forever on RED.
		select {
		case out := <-jobDone:
			resp.HTTPJobStatus = out.Status
			resp.HTTPJobOK = out.OK
			resp.HTTPJobError = out.Error
			resp.HTTPJobID = out.JobID
			resp.HTTPJobBody = out.Body
		case <-time.After(200 * time.Millisecond):
			// Job still waiting for result — expected for returns-queued-job.
		}
	}
	return resp, nil
}

func runResult(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ResultCase == "" {
		t.Fatal("ResultCase must be set by leaf Setup")
	}
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{ExitCode: 0, BaseURL: srv.BaseURL, ProbeURL: srv.BaseURL + "/v1/ext/result"}

	if req.ResultCase == ResultUnknown404 {
		body := map[string]any{
			"session_id": firstNonEmpty(req.ResultSessionID, req.SessionID, "no-such-session-ext-poll"),
			"job_id":     firstNonEmpty(req.ResultJobIDOverride, "job-missing"),
			"ok":         true,
		}
		status, ct, raw, err := doJSON(http.MethodPost, resp.ProbeURL, body, req.HTTPClientTimeout)
		if err != nil {
			return resp, err
		}
		resp.StatusCode = status
		resp.ContentType = ct
		resp.Body = raw
		resp.BodyString = string(raw)
		parseResultJSON(resp, raw)
		return resp, nil
	}

	// completes-ok: hello → enqueue → poll → result → collect job waiter.
	if err := postExtHello(srv.BaseURL, req); err != nil {
		return resp, fmt.Errorf("pre-hello: %w", err)
	}
	jobDone := startBackgroundJob(srv.BaseURL, req)
	time.Sleep(30 * time.Millisecond)

	pollBody := map[string]any{
		"session_id": req.SessionID,
		"wait_ms":    req.PollWaitMS,
	}
	if req.PollWaitMS <= 0 {
		pollBody["wait_ms"] = 2000
	}
	st, _, praw, err := doJSON(http.MethodPost, srv.BaseURL+"/v1/ext/poll", pollBody, 5*time.Second)
	if err != nil {
		return resp, fmt.Errorf("poll: %w", err)
	}
	if st != http.StatusOK {
		resp.StatusCode = st
		resp.BodyString = string(praw)
		return resp, fmt.Errorf("poll status=%d body=%s", st, praw)
	}
	parsePollJSON(resp, praw)
	jobID := resp.FirstJobID
	if jobID == "" {
		jobID = req.ResultJobIDOverride
	}
	if jobID == "" {
		return resp, fmt.Errorf("poll returned no job id; body=%s", praw)
	}

	rbody := map[string]any{
		"session_id": req.SessionID,
		"job_id":     jobID,
		"ok":         req.ResultOK,
	}
	if req.ResultData != nil {
		rbody["data"] = req.ResultData
	}
	if req.ResultError != "" {
		rbody["error"] = req.ResultError
	}
	status, ct, raw, err := doJSON(http.MethodPost, resp.ProbeURL, rbody, req.HTTPClientTimeout)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = raw
	resp.BodyString = string(raw)
	parseResultJSON(resp, raw)

	select {
	case out := <-jobDone:
		resp.HTTPJobStatus = out.Status
		resp.HTTPJobOK = out.OK
		resp.HTTPJobError = out.Error
		resp.HTTPJobID = out.JobID
		resp.HTTPJobBody = out.Body
	case <-time.After(jobWaitBudget(req)):
		resp.HTTPJobError = "timeout waiting for /v1/jobs after result"
	}
	return resp, nil
}

func runFlow(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.FlowCase == "" {
		t.Fatal("FlowCase must be set by leaf Setup")
	}
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{ExitCode: 0, BaseURL: srv.BaseURL}

	// 1) hello
	hstatus, _, hraw, err := doJSON(http.MethodPost, srv.BaseURL+"/v1/ext/hello", map[string]any{
		"session_id": req.SessionID,
		"version":    firstNonEmpty(req.HelloVersion, "1.0.0"),
		"features":   req.HelloFeatures,
	}, req.HTTPClientTimeout)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = hstatus
	resp.BodyString = string(hraw)
	parseHelloJSON(resp, hraw)
	if hstatus != http.StatusOK {
		return resp, nil
	}

	u := srv.BaseURL + "/v1/session?session=" + url.QueryEscape(req.SessionID)
	st, _, sbody, err := doJSON(http.MethodGet, u, nil, req.HTTPClientTimeout)
	if err != nil {
		return resp, err
	}
	resp.SessionProbeURL = u
	resp.SessionProbeStatus = st
	parseSessionConnected(resp, sbody)

	// 2) enqueue job in background
	jobDone := startBackgroundJob(srv.BaseURL, req)
	time.Sleep(30 * time.Millisecond)

	// 3) poll
	waitMS := req.PollWaitMS
	if waitMS <= 0 {
		waitMS = 3000
	}
	pstatus, _, praw, err := doJSON(http.MethodPost, srv.BaseURL+"/v1/ext/poll", map[string]any{
		"session_id": req.SessionID,
		"wait_ms":    waitMS,
	}, time.Duration(waitMS)*time.Millisecond+3*time.Second)
	if err != nil {
		return resp, err
	}
	// Keep hello status separate; final StatusCode reflects last critical step unless poll fails.
	if pstatus != http.StatusOK {
		resp.StatusCode = pstatus
		resp.BodyString = string(praw)
		return resp, nil
	}
	parsePollJSON(resp, praw)
	jobID := resp.FirstJobID
	if jobID == "" {
		resp.StatusCode = pstatus
		resp.BodyString = string(praw)
		return resp, fmt.Errorf("flow poll returned no jobs: %s", praw)
	}

	// 4) result
	rstatus, rct, rraw, err := doJSON(http.MethodPost, srv.BaseURL+"/v1/ext/result", map[string]any{
		"session_id": req.SessionID,
		"job_id":     jobID,
		"ok":         true,
		"data":       map[string]any{"flow": true},
	}, req.HTTPClientTimeout)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = rstatus
	resp.ContentType = rct
	resp.Body = rraw
	resp.BodyString = string(rraw)
	parseResultJSON(resp, rraw)

	select {
	case out := <-jobDone:
		resp.HTTPJobStatus = out.Status
		resp.HTTPJobOK = out.OK
		resp.HTTPJobError = out.Error
		resp.HTTPJobID = out.JobID
		resp.HTTPJobBody = out.Body
	case <-time.After(jobWaitBudget(req)):
		resp.HTTPJobError = "timeout waiting for /v1/jobs after flow result"
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
	if !req.OmitCreate && req.SessionID != "" {
		if _, err := reg.Create(req.SessionID); err != nil {
			return nil, nil, fmt.Errorf("registry Create %q: %w", req.SessionID, err)
		}
	}
	for _, id := range req.ExtraSessionIDs {
		if id == "" || id == req.SessionID {
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

// --- HTTP helpers ---

func doJSON(method, rawURL string, body any, timeout time.Duration) (status int, contentType string, raw []byte, err error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	var rdr io.Reader
	if body != nil {
		b, mErr := json.Marshal(body)
		if mErr != nil {
			return 0, "", nil, mErr
		}
		rdr = bytes.NewReader(b)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, rawURL, rdr)
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

func postExtHello(baseURL string, req *Request) error {
	body := map[string]any{
		"session_id": req.SessionID,
		"version":    firstNonEmpty(req.HelloVersion, "1.0.0"),
	}
	if req.HelloFeatures != nil {
		body["features"] = req.HelloFeatures
	} else {
		body["features"] = []string{"browser-agent"}
	}
	if req.HelloBundleMD5 != "" {
		body["bundle_md5"] = req.HelloBundleMD5
	}
	if req.HelloBrowser != "" {
		body["browser_product"] = req.HelloBrowser
	}
	st, _, raw, err := doJSON(http.MethodPost, baseURL+"/v1/ext/hello", body, req.HTTPClientTimeout)
	if err != nil {
		return err
	}
	if st != http.StatusOK {
		return fmt.Errorf("hello status=%d body=%s", st, raw)
	}
	return nil
}

type jobWaitOutcome struct {
	Status int
	OK     bool
	Error  string
	JobID  string
	Body   string
}

func startBackgroundJob(baseURL string, req *Request) chan jobWaitOutcome {
	ch := make(chan jobWaitOutcome, 1)
	jobType := req.JobHTTPType
	if jobType == "" {
		jobType = "eval"
	}
	params := req.JobHTTPParams
	if params == nil {
		params = map[string]any{"expression": "1+1"}
	}
	timeoutMS := req.JobHTTPTimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 8000
	}
	go func() {
		payload := map[string]any{
			"session_id": req.SessionID,
			"type":       jobType,
			"params":     params,
			"timeout_ms": timeoutMS,
		}
		clientTO := time.Duration(timeoutMS)*time.Millisecond + 2*time.Second
		st, _, raw, err := doJSON(http.MethodPost, baseURL+"/v1/jobs", payload, clientTO)
		out := jobWaitOutcome{Status: st, Body: string(raw)}
		if err != nil {
			out.Error = err.Error()
			ch <- out
			return
		}
		var m map[string]any
		if json.Unmarshal(raw, &m) == nil {
			if ok, _ := m["ok"].(bool); ok {
				out.OK = true
			}
			if e, _ := m["error"].(string); e != "" {
				out.Error = e
			}
			if id, _ := m["job_id"].(string); id != "" {
				out.JobID = id
			}
		}
		ch <- out
	}()
	return ch
}

func jobWaitBudget(req *Request) time.Duration {
	if req.JobWaitTimeout > 0 {
		return req.JobWaitTimeout
	}
	if req.JobHTTPTimeoutMS > 0 {
		return time.Duration(req.JobHTTPTimeoutMS)*time.Millisecond + time.Second
	}
	return 10 * time.Second
}

// --- parsers ---

func parseHelloJSON(resp *Response, raw []byte) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	resp.Raw = m
	if ok, _ := m["ok"].(bool); ok {
		resp.HTTPOK = true
	}
	if p, _ := m["phase"].(string); p != "" {
		resp.Phase = p
	}
}

func parseSessionConnected(resp *Response, raw []byte) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	if ext, ok := m["extension"].(map[string]any); ok {
		resp.ExtensionConnected, _ = ext["connected"].(bool)
	}
	// Some payloads expose top-level connected.
	if !resp.ExtensionConnected {
		if c, ok := m["connected"].(bool); ok {
			resp.ExtensionConnected = c
		}
	}
}

func parsePollJSON(resp *Response, raw []byte) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	resp.Raw = m
	resp.Jobs = asObjectSlice(m["jobs"])
	resp.Events = asObjectSlice(m["events"])
	resp.JobCount = len(resp.Jobs)
	resp.EventCount = len(resp.Events)
	if len(resp.Jobs) > 0 {
		j := resp.Jobs[0]
		resp.FirstJobID = stringField(j, "id", "job_id")
		resp.FirstJobType = stringField(j, "type")
	}
	for _, ev := range resp.Events {
		typ := stringField(ev, "type")
		if typ == "prepare_reconnect" {
			resp.HasPrepareReconnect = true
			break
		}
		// Nested envelope style: {type, payload} or payload-only with type inside.
		if p, ok := ev["payload"].(map[string]any); ok {
			if stringField(p, "type") == "prepare_reconnect" {
				resp.HasPrepareReconnect = true
				break
			}
		}
	}
}

func parseResultJSON(resp *Response, raw []byte) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	resp.Raw = m
	if ok, _ := m["ok"].(bool); ok {
		resp.ResultOKField = true
		resp.HTTPOK = true
	}
}

func asObjectSlice(v any) []map[string]any {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func stringField(m map[string]any, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, k := range keys {
		if s, ok := m[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// silence unused
var (
	_ = sync.Mutex{}
	_ = io.Discard
	_ = fmt.Sprintf
)
```
