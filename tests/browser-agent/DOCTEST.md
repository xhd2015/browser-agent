# browser-agent MVP — session resolve, job RPC, WS control, session UI

Exercises the **browser-agent** control plane (package
`github.com/xhd2015/browser-agent/browseragent`):

| Surface | What is under test |
|---------|-------------------|
| Session id resolve | flag vs env `BROWSER_AGENT_SESSION_ID`; neither → hard error |
| Job queue | in-memory FIFO, wait/complete, timeout → expired, unknown complete |
| HTTP jobs | `POST /v1/jobs` enqueue-and-wait; 404 unknown session |
| WebSocket control | hello → supports; job push; result unblocks; **disconnect fails inflight** |
| Session UI | `GET /v1/session` JSON + `GET /go` (or `/`) SPA shell / product markers |
| SYSTEM.md prompt | `FormatSystemPrompt` recipes + product default port **43761** |

**No real Chrome.** **No real agent-run.** Default path uses in-process Go
server (ephemeral `127.0.0.1:0`), pure package APIs, and a **fake extension**
WebSocket client in the harness.

This tree is **new** and independent of sealed browser-trace trees. After
implement, regressions stay green via:

```sh
doctest test ./tests/browser-trace
doctest test ./tests/browser-trace-session-page
doctest test ./tests/browser-trace-install-panel
```

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** (or **Agent**) talks to **browser-agent** through two channels:

1. **CLI / HTTP** — side commands under **`session`**
   (`browser-agent session info|eval|run|logs|screenshot|cdp`) resolve a
   **session id**, then `POST /v1/jobs` and **block until** a **JobResult**
   (or timeout). Flat `info`/`eval`/… are not side-command handlers.
2. **Session Page** — Vite SPA at `GET /` or `GET /go` polls `GET /v1/session`
   for connection / install guidance (product port **43761**).

**Control Server** (Go, one session per `serve`) keeps in-memory:

- **Session** — `id`, status/phase, extension connection, optional agent meta,
  **JobQueue**
- **Job** — `id`, `type` (`info|eval|run|logs|screenshot|cdp`), `params`,
  `timeout_ms`, status `queued|running|done|failed|expired`, optional result
- **JobResult** — `job_id`, `ok`, `error`, `data` (JSON), `duration_ms`

**Extension Agent** (Chrome MV3, product `Chrome-Ext-Browser-Agent`) connects
to `GET /v1/ws` and speaks a versioned envelope:

```text
{ v: 1, type: hello|status|job|result|cancel|ping|pong, id, payload }
```

- **hello** — announces `version` + `features[]` (must include `browser-agent`
  and version ≥ product floor for `supports_browser_agent`)
- **job** — server → extension push after enqueue
- **result** — extension → server; unblocks HTTP waiters
- **Disconnect policy (v1)**: if the WebSocket drops while jobs are
  queued/running, those **inflight jobs fail** immediately (no requeue).
  Documented and asserted in `ws-control/disconnect-fails-inflight`.

**Session id resolution** (CLI helpers / package):

1. `--session-id` / explicit flag value if provided  
2. else env `BROWSER_AGENT_SESSION_ID`  
3. else error mentioning **both** `--session-id` and `BROWSER_AGENT_SESSION_ID`

**SYSTEM.md** (written under session dir on serve; pure formatter also exported)
must teach the agent nested recipes **without** embedding a concrete control
session id:

```text
browser-agent session info
browser-agent session eval '…'
browser-agent session run path/to/script.js
browser-agent session logs
browser-agent session screenshot
```

Session resolve: `--session-id` or env `BROWSER_AGENT_SESSION_ID`.

**Product defaults**: control listen default **`127.0.0.1:43761`** (not 43759).
browser-trace remains 43759 (regression elsewhere; this tree asserts agent side).

**Test Client** in this tree:

- Pure leaves call `ResolveSessionID`, `JobQueue`, `FormatSystemPrompt`,
  product defaults — no listen socket.
- Integration leaves start `browseragent.Run` (or equivalent serve entry) with
  `NoOpenChrome`, no agent-run, temp `BaseDir`, known session id, free port;
  dial WS and/or HTTP; cancel context after the probe.

## Decision Tree

```
browser-agent
├── resolve-session/                         [pure ResolveSessionID]
│   ├── flag-wins/                             flag + env → flag
│   ├── env-only/                              env only → env
│   └── neither-error/                         neither → error mentions flag + env
├── job-queue/                               [pure JobQueue]
│   ├── enqueue-dequeue/
│   │   ├── single/                              FIFO identity
│   │   └── two-jobs-fifo/                       A then B
│   ├── complete-unblocks-waiter/              Wait returns Complete payload
│   ├── wait-timeout/                          no Complete → ok=false timeout
│   ├── complete-unknown-id/                   Complete missing id → error
│   └── expire-late-result/                    timeout → expired; late Complete safe
├── http-jobs/                               [POST /v1/jobs]
│   ├── known-session/
│   │   ├── extension-completes/                 fake WS result → 200 ok
│   │   └── no-extension-timeout/                short timeout → error timeout
│   └── unknown-session-404/                   wrong session → 404
├── ws-control/                              [GET /v1/ws + jobs]
│   ├── hello-supports/                        hello → session connected+supports
│   ├── job-push-on-enqueue/                   after hello, POST job → WS type=job
│   ├── result-unblocks-http/                  WS result unblocks HTTP waiter
│   └── disconnect-fails-inflight/             WS drop → inflight job failed
├── session-ui/                              [GET /v1/session + /go HTML]
│   ├── v1-session/
│   │   ├── no-hello-waiting/                    connected false
│   │   └── after-ws-hello-supports/             connected + supports_browser_agent
│   └── go-html/
│       ├── spa-shell-hooks/                     session id + /v1/session hooks
│       └── product-port-43761/                  product name/port markers
└── system-prompt/                           [FormatSystemPrompt + defaults]
    ├── format-contains-recipes/               nested session recipes; no control id
    └── product-defaults-port/                 default port 43761
```

### Parameter significance (high → low)

1. **Surface / Mode** — pure resolve vs job-queue vs HTTP vs WS vs session UI
   vs system prompt (different contracts; one Mode per top branch).
2. **Within surface** — success vs error / known vs unknown / connected vs not
   (largest outcome split).
3. **Leaf variants** — single vs two jobs, timeout vs complete, product string
   details.

## Test Index

| Leaf | Scenario |
|------|----------|
| `resolve-session/flag-wins` | (A1) flag + env → flag wins |
| `resolve-session/env-only` | (A2) env only → env value |
| `resolve-session/neither-error` | (A3) neither → error mentions `--session-id` and `BROWSER_AGENT_SESSION_ID` |
| `job-queue/enqueue-dequeue/single` | (B1) enqueue then dequeue — same id/type/params |
| `job-queue/enqueue-dequeue/two-jobs-fifo` | (B2) two jobs — A then B |
| `job-queue/complete-unblocks-waiter` | (B3) Complete unblocks Wait with equal payload |
| `job-queue/wait-timeout` | (B4) wait timeout, no complete → ok=false, error contains timeout |
| `job-queue/complete-unknown-id` | (B5) Complete unknown id → error |
| `job-queue/expire-late-result` | (B6) waiter timeout → status expired; late result safe/ignored |
| `http-jobs/known-session/extension-completes` | (C1) POST /v1/jobs; fake WS completes → 200 + ok |
| `http-jobs/known-session/no-extension-timeout` | (C2) POST /v1/jobs; no extension; short timeout → error timeout |
| `http-jobs/unknown-session-404` | (C3) unknown session → 404 |
| `ws-control/hello-supports` | (D1) WS hello feature `browser-agent` + version ≥ floor → supports |
| `ws-control/job-push-on-enqueue` | (D2) after hello, enqueue eval → WS `type=job` payload eval |
| `ws-control/result-unblocks-http` | (D3) extension `type=result` unblocks HTTP waiter |
| `ws-control/disconnect-fails-inflight` | (D4) disconnect mid-job → job fails (v1: no requeue) |
| `session-ui/v1-session/no-hello-waiting` | (E1) GET /v1/session known, no hello → waiting; connected false |
| `session-ui/v1-session/after-ws-hello-supports` | (E2) after WS hello → connected true; supports_browser_agent true |
| `session-ui/go-html/spa-shell-hooks` | (E3) GET /go or / HTML embeds session SPA hooks |
| `session-ui/go-html/product-port-43761` | (E4) install/product markers mention 43761 / browser-agent |
| `system-prompt/format-contains-recipes` | (F1) FormatSystemPrompt nested `session` recipes; no concrete control id; mentions `BROWSER_AGENT_SESSION_ID` |
| `system-prompt/product-defaults-port` | (F2) product default control port 43761 |

*(C4 missing session id is covered by `resolve-session/neither-error`. CLI binary
G leaves are deferred — package API is the GREEN path for MVP.)*

## How to Run

```sh
cd project-api-capture
doctest vet ./tests/browser-agent
doctest test ./tests/browser-agent/...
# or:
cd tests/browser-agent && doctest vet . && doctest test -v .
```

Requires package `github.com/xhd2015/browser-agent/browseragent`
(may not exist yet — leaves are **RED** until implementer lands APIs below).

### Expected package / wire contract (implementer)

Prose contract (authoritative for GREEN):

**Pure**

- `ResolveSessionID(flagValue string, flagSet bool, envValue string) (string, error)`  
  — or equivalent: flag wins when set; else non-empty env; else error text must
  include `--session-id` and `BROWSER_AGENT_SESSION_ID`.
- `NewJobQueue() *JobQueue` with `Enqueue`, `Dequeue`/`TryDequeue`, `Wait`,
  `Complete`, `Get` supporting statuses `queued|running|done|failed|expired`.
- `FormatSystemPrompt(sessionID string) string` — nested `browser-agent session …`
  recipes listed in DSN; must **not** embed the concrete control session id;
  must mention `BROWSER_AGENT_SESSION_ID`.
- Product default listen **`127.0.0.1:43761`** (const or `DefaultAddr` /
  `DefaultControlPort`).

**HTTP**

| Method | Path | Notes |
|--------|------|-------|
| GET | `/v1/health` | liveness 200 |
| GET | `/v1/session` | JSON snapshot (`?session=` optional if single-session serve) |
| POST | `/v1/jobs` | body `{session_id?, type, params, timeout_ms}`; **holds until result or timeout** |
| GET | `/v1/jobs/{id}` | optional poll |
| GET | `/` or `/go` | session SPA HTML |
| GET | `/v1/ws` | extension WebSocket |

**Session JSON** (fields used by asserts; extra fields OK):

```json
{
  "session_id": "...",
  "phase": "waiting_extension",
  "extension": {
    "connected": false,
    "version": "",
    "features": [],
    "supports_browser_agent": false
  },
  "hint": "..."
}
```

- `supports_browser_agent` true only when hello features include `browser-agent`
  **and** version ≥ product floor (recommend `1.0.0` unless product picks another).
- Unknown session id on job/session routes → **HTTP 404**.

**WS envelope** `v=1`; hello payload carries `version` + `features`; job push
`type=job`; result `type=result` with job id + ok/error/data.

**Serve entry** (name flexible): `Run(ctx, Config)` blocks until cancel; Config
includes `Addr`, `BaseDir`, `SessionID` (or suffix), `NoOpenChrome`, and a way
to skip agent-run / Chrome for tests.

**Disconnect v1**: fail inflight jobs on WS disconnect (status `failed`, error
mentions disconnect or connection lost).

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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode values — top-level surface under test.
const (
	ModeResolveSession = "resolve-session"
	ModeJobQueue       = "job-queue"
	ModeHTTPJobs       = "http-jobs"
	ModeWSControl      = "ws-control"
	ModeSessionUI      = "session-ui"
	ModeSystemPrompt   = "system-prompt"
)

// JobOp values for ModeJobQueue.
const (
	JobOpEnqueueDequeue   = "enqueue-dequeue"
	JobOpFIFOTwo          = "fifo-two"
	JobOpCompleteUnblocks = "complete-unblocks"
	JobOpWaitTimeout      = "wait-timeout"
	JobOpCompleteUnknown  = "complete-unknown"
	JobOpExpireLate       = "expire-late"
)

// SystemOp values for ModeSystemPrompt.
const (
	SystemOpFormat   = "format"
	SystemOpDefaults = "product-defaults"
)

// Probe values for ModeSessionUI.
const (
	ProbeV1Session = "v1-session"
	ProbeGoHTML    = "go-html"
)

// WSAction values for ModeWSControl.
const (
	WSActionHelloSupports     = "hello-supports"
	WSActionJobPush           = "job-push"
	WSActionResultUnblocks    = "result-unblocks"
	WSActionDisconnectInflight = "disconnect-inflight"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// --- resolve-session ---
	FlagSet   bool
	FlagValue string
	EnvSet    bool
	EnvValue  string

	// --- job-queue ---
	JobOp         string
	JobType       string
	JobParams     map[string]any
	JobTimeout    time.Duration
	SecondJobType string
	SecondParams  map[string]any
	// CompleteData is JSON-ish payload for Complete (complete-unblocks).
	CompleteOK    bool
	CompleteError string
	CompleteData  map[string]any

	// --- server common (http-jobs, ws-control, session-ui) ---
	Addr         string
	BaseDir      string
	SessionID    string
	NoOpenChrome bool
	// ReadyTimeout bounds health wait / short serve windows.
	ReadyTimeout time.Duration
	// JobHTTPTimeoutMS is timeout_ms on POST /v1/jobs (0 → Run default).
	JobHTTPTimeoutMS int64
	// JobHTTPType / JobHTTPParams for POST /v1/jobs.
	JobHTTPType   string
	JobHTTPParams map[string]any
	// ForceUnknownSession probes with a non-live session id.
	ForceUnknownSession bool
	SessionIDForProbe   string

	// HTTPJobs: whether harness connects a fake extension and completes the job.
	FakeExtension       bool
	FakeExtensionResult bool // if FakeExtension, send ok result for the first job

	// --- ws-control ---
	WSAction      string
	HelloVersion  string
	HelloFeatures []string

	// --- session-ui ---
	Probe     string
	DoWSHello bool

	// --- system-prompt ---
	SystemOp        string
	PromptSessionID string
}

// Response holds pure + HTTP/WS outcomes.
type Response struct {
	// ResolveSession
	ResolvedID string
	ResolveErr string

	// Job queue
	DequeuedIDs     []string
	DequeuedTypes   []string
	JobStatus       string
	JobResultOK     bool
	JobResultError  string
	JobResultData   map[string]any
	CompleteErr     string
	LateCompleteErr string
	// JobGetStatus is status after Wait/timeout (Get).
	JobGetStatus string

	// HTTP
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string
	// Parsed job result from POST /v1/jobs when JSON.
	HTTPJobOK    bool
	HTTPJobError string
	HTTPJobID    string
	Raw          map[string]any

	// Session JSON
	SessionIDField           string
	Phase                    string
	ExtensionConnected       bool
	ExtensionVersion         string
	ExtensionFeatures        []string
	SupportsBrowserAgent     bool
	Hint                     string

	// WS
	WSHelloOK       bool
	WSJobReceived   bool
	WSJobType       string
	WSJobPayloadRaw string
	WSDisconnected  bool

	// System prompt / product
	SystemPrompt   string
	DefaultAddr    string
	DefaultPort    string
	ProductName    string

	// Meta
	RealSessionID string
	BaseURL       string
	ProbeURL      string
	RunErrText    string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeResolveSession:
		return runResolveSession(t, req)
	case ModeJobQueue:
		return runJobQueue(t, req)
	case ModeHTTPJobs:
		return runHTTPJobs(t, req)
	case ModeWSControl:
		return runWSControl(t, req)
	case ModeSessionUI:
		return runSessionUI(t, req)
	case ModeSystemPrompt:
		return runSystemPrompt(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runResolveSession(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	id, err := browseragent.ResolveSessionID(req.FlagValue, req.FlagSet, req.EnvValue, req.EnvSet)
	resp := &Response{ResolvedID: id}
	if err != nil {
		resp.ResolveErr = err.Error()
	}
	return resp, nil
}

func runJobQueue(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.JobOp == "" {
		t.Fatal("JobOp must be set by Setup")
	}
	q := browseragent.NewJobQueue()
	resp := &Response{}

	switch req.JobOp {
	case JobOpEnqueueDequeue:
		j := browseragent.Job{
			Type:   req.JobType,
			Params: req.JobParams,
		}
		if j.Type == "" {
			j.Type = "eval"
		}
		enqueued, err := q.Enqueue(j)
		if err != nil {
			return resp, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		got, err := q.Dequeue(ctx)
		if err != nil {
			return resp, err
		}
		resp.DequeuedIDs = []string{got.ID}
		resp.DequeuedTypes = []string{got.Type}
		resp.JobStatus = got.Status
		if got.ID != enqueued.ID {
			return resp, fmt.Errorf("dequeued id %q != enqueued %q", got.ID, enqueued.ID)
		}
		// Params round-trip checked in Assert via JobResultData snapshot.
		resp.JobResultData = map[string]any{"params": got.Params, "type": got.Type, "id": got.ID}

	case JobOpFIFOTwo:
		aType := req.JobType
		if aType == "" {
			aType = "eval"
		}
		bType := req.SecondJobType
		if bType == "" {
			bType = "info"
		}
		a, err := q.Enqueue(browseragent.Job{Type: aType, Params: req.JobParams})
		if err != nil {
			return resp, err
		}
		b, err := q.Enqueue(browseragent.Job{Type: bType, Params: req.SecondParams})
		if err != nil {
			return resp, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		first, err := q.Dequeue(ctx)
		if err != nil {
			return resp, err
		}
		second, err := q.Dequeue(ctx)
		if err != nil {
			return resp, err
		}
		resp.DequeuedIDs = []string{first.ID, second.ID}
		resp.DequeuedTypes = []string{first.Type, second.Type}
		_ = a
		_ = b

	case JobOpCompleteUnblocks:
		j, err := q.Enqueue(browseragent.Job{Type: "eval", Params: map[string]any{"code": "1+1"}})
		if err != nil {
			return resp, err
		}
		waitDone := make(chan struct{})
		var waitRes browseragent.JobResult
		var waitErr error
		go func() {
			defer close(waitDone)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			waitRes, waitErr = q.Wait(ctx, j.ID)
		}()
		// Brief yield so Wait is registered.
		time.Sleep(20 * time.Millisecond)
		data := req.CompleteData
		if data == nil {
			data = map[string]any{"value": 2}
		}
		ok := req.CompleteOK
		if !req.CompleteOK && req.CompleteError == "" {
			ok = true
		}
		cerr := q.Complete(j.ID, browseragent.JobResult{
			JobID: j.ID,
			OK:    ok,
			Error: req.CompleteError,
			Data:  data,
		})
		if cerr != nil {
			resp.CompleteErr = cerr.Error()
		}
		select {
		case <-waitDone:
		case <-time.After(3 * time.Second):
			return resp, fmt.Errorf("Wait did not unblock after Complete")
		}
		if waitErr != nil {
			resp.JobResultError = waitErr.Error()
		}
		resp.JobResultOK = waitRes.OK
		resp.JobResultError = firstNonEmpty(resp.JobResultError, waitRes.Error)
		resp.JobResultData = waitRes.Data
		if got, ok := q.Get(j.ID); ok {
			resp.JobGetStatus = got.Status
		}

	case JobOpWaitTimeout:
		timeout := req.JobTimeout
		if timeout <= 0 {
			timeout = 80 * time.Millisecond
		}
		j, err := q.Enqueue(browseragent.Job{
			Type:      "eval",
			TimeoutMS: timeout.Milliseconds(),
		})
		if err != nil {
			return resp, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		res, werr := q.Wait(ctx, j.ID)
		if werr != nil {
			resp.JobResultError = werr.Error()
		}
		resp.JobResultOK = res.OK
		if res.Error != "" {
			resp.JobResultError = firstNonEmpty(resp.JobResultError, res.Error)
		}
		if got, ok := q.Get(j.ID); ok {
			resp.JobGetStatus = got.Status
		}

	case JobOpCompleteUnknown:
		err := q.Complete("no-such-job-id", browseragent.JobResult{OK: true})
		if err != nil {
			resp.CompleteErr = err.Error()
		}

	case JobOpExpireLate:
		timeout := req.JobTimeout
		if timeout <= 0 {
			timeout = 80 * time.Millisecond
		}
		j, err := q.Enqueue(browseragent.Job{
			Type:      "screenshot",
			TimeoutMS: timeout.Milliseconds(),
		})
		if err != nil {
			return resp, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		res, werr := q.Wait(ctx, j.ID)
		cancel()
		if werr != nil {
			resp.JobResultError = werr.Error()
		}
		resp.JobResultOK = res.OK
		if res.Error != "" {
			resp.JobResultError = firstNonEmpty(resp.JobResultError, res.Error)
		}
		if got, ok := q.Get(j.ID); ok {
			resp.JobGetStatus = got.Status
		}
		// Late complete must not panic; may error or be ignored.
		lateErr := q.Complete(j.ID, browseragent.JobResult{
			JobID: j.ID,
			OK:    true,
			Data:  map[string]any{"late": true},
		})
		if lateErr != nil {
			resp.LateCompleteErr = lateErr.Error()
		}
		if got, ok := q.Get(j.ID); ok {
			// Prefer status after late complete (should stay expired/failed, not flip to done via late ok).
			resp.JobGetStatus = got.Status
		}

	default:
		return nil, fmt.Errorf("unknown JobOp %q", req.JobOp)
	}
	return resp, nil
}

func runHTTPJobs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startAgentServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		RealSessionID: srv.SessionID,
		BaseURL:       srv.BaseURL,
	}

	sessionForPost := srv.SessionID
	if req.ForceUnknownSession {
		sessionForPost = req.SessionIDForProbe
		if sessionForPost == "" {
			sessionForPost = "does-not-exist"
		}
	}

	// Optional fake extension that completes the first job it receives.
	var ext *fakeExtension
	if req.FakeExtension {
		ext, err = dialFakeExtension(srv.BaseURL, req.HelloVersion, req.HelloFeatures)
		if err != nil {
			return resp, fmt.Errorf("fake extension dial: %w", err)
		}
		defer ext.Close()
		if req.FakeExtensionResult {
			ext.AutoCompleteOK = true
			go ext.Loop()
			// Allow hello to settle.
			time.Sleep(30 * time.Millisecond)
		}
	}

	timeoutMS := req.JobHTTPTimeoutMS
	if timeoutMS <= 0 {
		if req.FakeExtension && req.FakeExtensionResult {
			timeoutMS = 3000
		} else {
			timeoutMS = 150
		}
	}
	jobType := req.JobHTTPType
	if jobType == "" {
		jobType = "eval"
	}
	params := req.JobHTTPParams
	if params == nil {
		params = map[string]any{"code": "1+1"}
	}

	status, ct, body, postErr := postJobs(srv.BaseURL, sessionForPost, jobType, params, timeoutMS)
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = body
	resp.BodyString = string(body)
	resp.ProbeURL = srv.BaseURL + "/v1/jobs"
	if postErr != nil {
		// Transport error — surface to Assert.
		return resp, postErr
	}
	parseJobHTTPResult(resp, body)
	return resp, nil
}

func runWSControl(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.WSAction == "" {
		t.Fatal("WSAction must be set by Setup")
	}
	srv, cleanup, err := startAgentServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		RealSessionID: srv.SessionID,
		BaseURL:       srv.BaseURL,
	}
	version := req.HelloVersion
	if version == "" {
		version = "1.0.0"
	}
	features := req.HelloFeatures
	if features == nil {
		features = []string{"browser-agent"}
	}

	switch req.WSAction {
	case WSActionHelloSupports:
		ext, err := dialFakeExtension(srv.BaseURL, version, features)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		if err := ext.SendHello(); err != nil {
			return resp, err
		}
		resp.WSHelloOK = true
		time.Sleep(40 * time.Millisecond)
		if err := fillSessionProbe(resp, srv.BaseURL, srv.SessionID); err != nil {
			return resp, err
		}

	case WSActionJobPush:
		ext, err := dialFakeExtension(srv.BaseURL, version, features)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		if err := ext.SendHello(); err != nil {
			return resp, err
		}
		resp.WSHelloOK = true
		time.Sleep(30 * time.Millisecond)

		jobCh := make(chan wsEnvelope, 1)
		ext.OnJob = func(env wsEnvelope) {
			select {
			case jobCh <- env:
			default:
			}
		}
		go ext.Loop()

		// Fire-and-forget POST in background with long timeout; we only need WS push.
		go func() {
			_, _, _, _ = postJobs(srv.BaseURL, srv.SessionID, "eval", map[string]any{"code": "1"}, 5000)
		}()

		select {
		case env := <-jobCh:
			resp.WSJobReceived = true
			resp.WSJobType = envelopeJobType(env)
			b, _ := json.Marshal(env)
			resp.WSJobPayloadRaw = string(b)
		case <-time.After(3 * time.Second):
			return resp, fmt.Errorf("timed out waiting for WS job push")
		}

	case WSActionResultUnblocks:
		ext, err := dialFakeExtension(srv.BaseURL, version, features)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		if err := ext.SendHello(); err != nil {
			return resp, err
		}
		ext.AutoCompleteOK = true
		go ext.Loop()
		time.Sleep(30 * time.Millisecond)

		status, ct, body, postErr := postJobs(srv.BaseURL, srv.SessionID, "eval", map[string]any{"code": "2+2"}, 3000)
		resp.StatusCode = status
		resp.ContentType = ct
		resp.Body = body
		resp.BodyString = string(body)
		if postErr != nil {
			return resp, postErr
		}
		parseJobHTTPResult(resp, body)
		resp.WSJobReceived = ext.JobsSeen > 0

	case WSActionDisconnectInflight:
		ext, err := dialFakeExtension(srv.BaseURL, version, features)
		if err != nil {
			return resp, err
		}
		if err := ext.SendHello(); err != nil {
			ext.Close()
			return resp, err
		}
		// Do not auto-complete; disconnect after job is seen (or shortly after POST starts).
		jobSeen := make(chan struct{}, 1)
		ext.OnJob = func(env wsEnvelope) {
			select {
			case jobSeen <- struct{}{}:
			default:
			}
		}
		go ext.Loop()
		time.Sleep(30 * time.Millisecond)

		type postOut struct {
			status int
			ct     string
			body   []byte
			err    error
		}
		done := make(chan postOut, 1)
		go func() {
			// Longer client wait so disconnect path can fail the job first.
			st, ct, body, err := postJobs(srv.BaseURL, srv.SessionID, "eval", map[string]any{"code": "hang"}, 5000)
			done <- postOut{st, ct, body, err}
		}()

		// Wait until job pushed or brief fallback, then drop WS.
		select {
		case <-jobSeen:
		case <-time.After(1 * time.Second):
		}
		ext.Close()
		resp.WSDisconnected = true

		select {
		case out := <-done:
			resp.StatusCode = out.status
			resp.ContentType = out.ct
			resp.Body = out.body
			resp.BodyString = string(out.body)
			if out.err != nil {
				return resp, out.err
			}
			parseJobHTTPResult(resp, out.body)
		case <-time.After(6 * time.Second):
			return resp, fmt.Errorf("POST /v1/jobs did not return after disconnect")
		}

	default:
		return nil, fmt.Errorf("unknown WSAction %q", req.WSAction)
	}
	return resp, nil
}

func runSessionUI(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Probe == "" {
		t.Fatal("Probe must be set by Setup")
	}
	srv, cleanup, err := startAgentServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		RealSessionID: srv.SessionID,
		BaseURL:       srv.BaseURL,
	}

	if req.DoWSHello {
		version := req.HelloVersion
		if version == "" {
			version = "1.0.0"
		}
		features := req.HelloFeatures
		if features == nil {
			features = []string{"browser-agent"}
		}
		ext, err := dialFakeExtension(srv.BaseURL, version, features)
		if err != nil {
			return resp, err
		}
		defer ext.Close()
		if err := ext.SendHello(); err != nil {
			return resp, err
		}
		resp.WSHelloOK = true
		time.Sleep(40 * time.Millisecond)
	}

	switch req.Probe {
	case ProbeV1Session:
		if err := fillSessionProbe(resp, srv.BaseURL, srv.SessionID); err != nil {
			return resp, err
		}
	case ProbeGoHTML:
		// Prefer /go; fall back to / if product only mounts root.
		u := srv.BaseURL + "/go"
		if req.SessionID != "" {
			u = u + "?session=" + url.QueryEscape(srv.SessionID)
		}
		status, ct, body, err := doGET(u)
		if err != nil {
			return resp, err
		}
		// If /go is missing, try /.
		if status == http.StatusNotFound {
			u2 := srv.BaseURL + "/"
			status, ct, body, err = doGET(u2)
			if err != nil {
				return resp, err
			}
			u = u2
		}
		resp.StatusCode = status
		resp.ContentType = ct
		resp.Body = body
		resp.BodyString = string(body)
		resp.ProbeURL = u
	default:
		return nil, fmt.Errorf("unknown Probe %q", req.Probe)
	}
	return resp, nil
}

func runSystemPrompt(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SystemOp == "" {
		t.Fatal("SystemOp must be set by Setup")
	}
	resp := &Response{}
	switch req.SystemOp {
	case SystemOpFormat:
		sid := req.PromptSessionID
		if sid == "" {
			sid = "sess-system-prompt"
		}
		resp.SystemPrompt = browseragent.FormatSystemPrompt(sid)
		resp.SessionIDField = sid
	case SystemOpDefaults:
		// Prefer exported defaults; accept several naming styles via helpers.
		resp.DefaultAddr = browseragent.DefaultAddr
		if p := portFromAddr(resp.DefaultAddr); p != "" {
			resp.DefaultPort = p
		}
		if resp.DefaultPort == "" {
			resp.DefaultPort = browseragent.DefaultControlPort
		}
		resp.ProductName = "browser-agent"
	default:
		return nil, fmt.Errorf("unknown SystemOp %q", req.SystemOp)
	}
	return resp, nil
}

// --- server harness ---

type agentServer struct {
	BaseURL   string
	SessionID string
	cancel    context.CancelFunc
	done      <-chan error
}

func startAgentServer(t *testing.T, req *Request) (*agentServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	if req.SessionID == "" {
		t.Fatal("SessionID must be set by Setup")
	}
	addr := req.Addr
	if addr == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, nil, err
		}
		addr = ln.Addr().String()
		_ = ln.Close()
		req.Addr = addr
	}
	ready := req.ReadyTimeout
	if ready <= 0 {
		ready = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	var stdout, stderr bytes.Buffer
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      req.BaseDir,
		SessionID:    req.SessionID,
		NoOpenChrome: true,
		NoAgentRun:   true,
		Stdout:       &stdout,
		Stderr:       &stderr,
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
	// Client timeout slightly above server job timeout.
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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

func parseJobHTTPResult(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.Raw = raw
	// Accept either flat JobResult or wrapped {result:{...}}.
	src := raw
	if nested, ok := raw["result"].(map[string]any); ok {
		src = nested
	}
	if id, ok := src["job_id"].(string); ok {
		resp.HTTPJobID = id
	} else if id, ok := src["id"].(string); ok {
		resp.HTTPJobID = id
	}
	if ok, exists := src["ok"].(bool); exists {
		resp.HTTPJobOK = ok
	}
	if e, ok := src["error"].(string); ok {
		resp.HTTPJobError = e
	}
	// Also allow top-level error string on non-result errors.
	if resp.HTTPJobError == "" {
		if e, ok := raw["error"].(string); ok {
			resp.HTTPJobError = e
		}
	}
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
		resp.ExtensionVersion, _ = ext["version"].(string)
		if v, ok := ext["supports_browser_agent"].(bool); ok {
			resp.SupportsBrowserAgent = v
		}
		if feats, ok := ext["features"].([]any); ok {
			for _, f := range feats {
				if s, ok := f.(string); ok {
					resp.ExtensionFeatures = append(resp.ExtensionFeatures, s)
				}
			}
		}
	}
}

// --- fake extension WS client ---

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
	return &fakeExtension{
		conn:     conn,
		version:  version,
		features: features,
	}, nil
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

func portFromAddr(addr string) string {
	if addr == "" {
		return ""
	}
	// host:port or bare port
	if i := strings.LastIndex(addr, ":"); i >= 0 && i+1 < len(addr) {
		return addr[i+1:]
	}
	return addr
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// Silence unused import guards when some modes dominate a compile unit.
var (
	_ = sync.Mutex{}
	_ = io.Discard
)
```
