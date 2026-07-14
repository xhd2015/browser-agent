# browser-trace session page — live status JSON + HTML

Exercises the **session page status surface** of the browser-trace control
server: `GET /v1/session` (JSON for the page poller) and `GET /go` (HTML
dashboard with status UI hooks). Tests drive the control server over HTTP with
an in-process client (hello / status posts as needed). **No real Chrome** and
no real extension process.

This tree is **separate** from `./tests/browser-trace/` (sealed lifecycle /
HAR / deadlines). Existing six lifecycle leaves stay untouched.

Product defaults relevant here:

| Setting | Default |
|---------|---------|
| listen address | `127.0.0.1:43759` (tests use ephemeral free port) |
| ready timeout | `30s` (tests use short windows so probes stay fast) |
| capability rule | features must contain `browser-trace` **and** version ≥ `1.2.0` |
| page poll interval | ~500ms–1s (HTML must reference `/v1/session`) |

## Version

0.0.2

# DSN (Domain Specific Notion)

**User** opens the **Session Page** (`GET /go?session=<id>`) in a Chrome window
owned by a browser-trace session. The page is served by the **Control Server**.

**Control Server** keeps one in-memory **Session** and exposes:

- `GET /go?session=…` — HTML dashboard: session id, status UI root, inline JS
  that polls session JSON
- `GET /v1/session?session=…` — JSON snapshot for the poller (phase, extension
  connection/capability, recording counters, ready countdown, actionable hint)
- `POST /v1/hello` — **Extension Agent** announces `version` + `features[]`
- `POST /v1/status` — agent heartbeat (`state`, `entry_count`, `window_id`)
- (other wire paths exist for capture lifecycle; this tree only needs hello/status
  to stage session state before probing status endpoints)

**Session phase** (UI-facing):

- `waiting_extension` — no successful hello yet
- `extension_connected` — hello OK, not yet recording
- `recording` — status reported recording (capture active)
- `stopping` / `saved` / `failed` — later lifecycle (not primary leaves here)

**Capability gate** (`supports_browser_trace`):

- `true` only when hello features include `browser-trace` **and** version ≥ `1.2.0`
  (semver-ish major.minor.patch)
- version alone (no features) is **not** enough → `supports_browser_trace=false`
- connected can be true while support is false (agent present but too old / missing feature)

**Page poller** (inline JS on `/go`) reads `/v1/session` same-origin; server JSON
is authoritative for agent connected. Optional content-script marker
`window.__BROWSER_TRACE_EXT__` is a client-side hint only (not asserted here).

**Test Client** in this tree starts the control server (via `browsertrace.Run` with
`NoOpenChrome` and a known `SessionSuffix` as session id), optionally posts hello
and/or recording status, then probes `/v1/session` or `/go` while the session is live.

## Decision Tree

```
browser-trace-session-page (live status surface)
├── v1-session/                              [GET /v1/session JSON]
│   ├── session-missing/                       unknown session id → HTTP 404 + not-found body
│   └── session-known/                         real session id
│       ├── no-hello/                            waiting; connected=false; supports=false; hint waiting/install
│       └── hello-connected/                   after POST /v1/hello
│           ├── supports-browser-trace/          feature browser-trace + version ≥ 1.2.0
│           │   ├── not-yet-recording/             connected; supports=true; phase extension_connected
│           │   └── recording/                     phase recording; active; entry_count reflected
│           └── unsupported/                     connected=true; supports=false; hint update/support
│               ├── missing-feature/               version ≥ 1.2.0 but features omit browser-trace
│               ├── version-too-low/               features include browser-trace; version < 1.2.0
│               └── version-only-no-features/      hello body has version only (features required)
└── go-page/                                 [GET /go HTML]
    └── valid-session-html/                    session id + status UI root + /v1/session poll reference
```

### Parameter significance (high → low)

1. **HTTP surface** — `/v1/session` JSON vs `/go` HTML (different contracts).
2. **Session identity** — known id vs unknown (404 path).
3. **Extension connection** — no hello vs hello received.
4. **Capability** — supports browser-trace vs unsupported (when connected).
5. **Recording phase** — not yet recording vs status `recording` + entry_count
   (only meaningful once support is established for the primary happy path).
6. **Unsupported reason** — missing feature / version too low / version-only body
   (same supports=false outcome; different hello inputs).

## Test Index

| Leaf | Scenario (requirement #) |
|------|--------------------------|
| `v1-session/session-known/no-hello` | (#1) No hello → waiting; connected=false; supports=false; non-empty waiting/install hint |
| `v1-session/session-known/hello-connected/supports-browser-trace/not-yet-recording` | (#2) Hello with feature + version ≥ 1.2.0 → connected; supports=true; version echoed |
| `v1-session/session-known/hello-connected/unsupported/missing-feature` | (#3a) Hello without `browser-trace` feature → supports=false; update/support hint |
| `v1-session/session-known/hello-connected/unsupported/version-too-low` | (#3b) Hello with feature but version &lt; 1.2.0 → supports=false |
| `v1-session/session-known/hello-connected/unsupported/version-only-no-features` | (#3c) Version-only hello → supports=false (features required; version alone not enough) |
| `v1-session/session-known/hello-connected/supports-browser-trace/recording` | (#4) After status recording → phase recording; active; entry_count |
| `v1-session/session-missing` | (#5) Unknown session id → HTTP 404; body indicates not found |
| `go-page/valid-session-html` | (#6) GET `/go?session=valid` HTML has session id, status UI root, `/v1/session` poll reference |

## How to Run

```sh
cd tests/browser-trace-session-page
doctest vet .
doctest test -v .
# or from repo root:
doctest vet ./tests/browser-trace-session-page
doctest test ./tests/browser-trace-session-page
# regression (lifecycle tree must stay green):
doctest test ./tests/browser-trace
```

Requires package `github.com/xhd2015/browser-agent/browsertrace`.
Leaves fail at runtime until the implementer adds `GET /v1/session`, capability
fields on hello/session state, and status UI hooks on `/go` (TDD red → green).

### Expected control-server contract (implementer)

Prose contract (not a harness code block):

- `POST /v1/hello` accepts JSON `{"version":"…","features":["browser-trace",…]}`.
  `supports_browser_trace` is true **only if** `features` contains `browser-trace`
  **and** `version` ≥ `1.2.0`. Version without features → support false.
- `GET /v1/session?session=<id>` returns JSON:

```json
{
  "session_id": "...",
  "phase": "waiting_extension",
  "extension": {
    "connected": false,
    "version": "",
    "features": [],
    "supports_browser_trace": false
  },
  "recording": {
    "active": false,
    "entry_count": 0,
    "window_id": 0
  },
  "ready": {
    "deadline_ms": 30000,
    "elapsed_ms": 0,
    "remaining_ms": 30000
  },
  "hint": "Waiting for API Capture extension…"
}
```

- Unknown non-empty session id → **HTTP 404** JSON error indicating not found.
- `GET /go?session=<id>` HTML includes the session id, a stable status UI root
  (e.g. `id` or `data-browser-trace-status` on a root element), and client JS that
  references `/v1/session` for polling.
- Harness starts a live session via `browsertrace.Run` with `NoOpenChrome=true`,
  ephemeral `Addr`, temp `BaseDir`, and fixed `SessionSuffix` as the session id;
  probes run **while** the server is still accepting HTTP (before ready/complete end).

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browsertrace"
)

// Probe targets (Request.Probe).
const (
	ProbeV1Session = "v1-session"
	ProbeGoHTML    = "go"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Addr is host:port for Control Server. Empty → free loopback port in Run.
	Addr string
	// BaseDir is session parent directory (temp per leaf).
	BaseDir string
	// ReadyTimeout / CompleteTimeout override product defaults for fast tests.
	ReadyTimeout    time.Duration
	CompleteTimeout time.Duration
	// NoOpenChrome skips launching Chrome (always true in this tree).
	NoOpenChrome bool
	// SessionSuffix is the known session id (browsertrace uses suffix as id).
	SessionSuffix string

	// Probe selects which HTTP surface Run hits after staging.
	// ProbeV1Session or ProbeGoHTML.
	Probe string

	// ForceUnknownSession, when true, probes with a session id that does not
	// match the live session (404 path). SessionIDForProbe may override the
	// bogus id; default is "does-not-exist".
	ForceUnknownSession bool
	SessionIDForProbe   string

	// DoHello, when true, POST /v1/hello before the probe.
	DoHello bool
	// HelloVersion is the version field on hello (e.g. "1.2.0").
	HelloVersion string
	// HelloFeatures is the features array. Ignored when HelloOmitFeatures.
	HelloFeatures []string
	// HelloOmitFeatures sends {"version":…} only (no features key).
	HelloOmitFeatures bool

	// DoStatusRecording, when true, POST /v1/status with state=recording
	// after hello (and after a short settle). Requires DoHello for a realistic path.
	DoStatusRecording bool
	// EntryCount / WindowID reported in status (defaults applied in Run).
	EntryCount int
	WindowID   int
}

// Response holds the HTTP probe result (and light run metadata).
type Response struct {
	// StatusCode is the HTTP status from the probe.
	StatusCode int
	// ContentType is the Content-Type response header.
	ContentType string
	// Body is the raw response body.
	Body []byte
	// BodyString is string(Body) for convenience.
	BodyString string

	// Parsed session JSON fields (when probe is v1-session and body is JSON).
	SessionID              string
	Phase                  string
	ExtensionConnected     bool
	ExtensionVersion       string
	ExtensionFeatures      []string
	SupportsBrowserTrace   bool
	RecordingActive        bool
	EntryCount             int
	WindowID               int
	ReadyDeadlineMS        int64
	ReadyElapsedMS         int64
	ReadyRemainingMS       int64
	Hint                   string
	// Raw is the full decoded JSON object when parse succeeds.
	Raw map[string]any

	// RealSessionID is the live session id used by the control server.
	RealSessionID string
	// BaseURL is http://Addr.
	BaseURL string
	// ProbeURL is the full URL that was requested.
	ProbeURL string

	// RunExitCode / RunErrText from browsertrace.Run after cancel/timeout
	// (informational; leaves assert the probe, not full lifecycle exit).
	RunExitCode int
	RunErrText  string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	if req.SessionSuffix == "" {
		t.Fatal("SessionSuffix must be set by Setup (known session id)")
	}
	if req.Probe == "" {
		t.Fatal("Probe must be set by Setup")
	}
	req.NoOpenChrome = true
	if req.ReadyTimeout <= 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout <= 0 {
		req.CompleteTimeout = 2 * time.Second
	}
	if req.WindowID == 0 {
		req.WindowID = 42
	}

	addr := req.Addr
	if addr == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}
		addr = ln.Addr().String()
		_ = ln.Close()
		req.Addr = addr
	}
	baseURL := "http://" + addr
	realSessionID := req.SessionSuffix

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()

	var stdout, stderr bytes.Buffer
	cfg := browsertrace.Config{
		Addr:            addr,
		BaseDir:         req.BaseDir,
		ReadyTimeout:    req.ReadyTimeout,
		CompleteTimeout: req.CompleteTimeout,
		NoOpenChrome:    true,
		SessionSuffix:   req.SessionSuffix,
		Stdout:          &stdout,
		Stderr:          &stderr,
	}

	type runOutcome struct {
		result *browsertrace.Result
		err    error
	}
	runDone := make(chan runOutcome, 1)
	go func() {
		res, err := browsertrace.Run(runCtx, cfg)
		runDone <- runOutcome{result: res, err: err}
	}()

	// Wait until control server answers health (or ready window expires).
	if err := waitHealth(baseURL, req.ReadyTimeout); err != nil {
		runCancel()
		<-runDone
		return nil, fmt.Errorf("control server never became healthy at %s: %w", baseURL, err)
	}

	// Stage extension state for the probe.
	if req.DoHello {
		if err := postHello(baseURL, req.HelloVersion, req.HelloFeatures, req.HelloOmitFeatures); err != nil {
			runCancel()
			<-runDone
			return nil, fmt.Errorf("POST /v1/hello: %w", err)
		}
		// Brief settle so session state updates before probe/status.
		time.Sleep(30 * time.Millisecond)
	}
	if req.DoStatusRecording {
		if err := postStatusRecording(baseURL, req.EntryCount, req.WindowID); err != nil {
			runCancel()
			<-runDone
			return nil, fmt.Errorf("POST /v1/status recording: %w", err)
		}
		time.Sleep(30 * time.Millisecond)
	}

	probeSessionID := realSessionID
	if req.ForceUnknownSession {
		probeSessionID = req.SessionIDForProbe
		if probeSessionID == "" {
			probeSessionID = "does-not-exist"
		}
	} else if req.SessionIDForProbe != "" {
		probeSessionID = req.SessionIDForProbe
	}

	var probeURL string
	switch req.Probe {
	case ProbeV1Session:
		probeURL = baseURL + "/v1/session?session=" + probeSessionID
	case ProbeGoHTML:
		probeURL = baseURL + "/go?session=" + probeSessionID
	default:
		runCancel()
		<-runDone
		return nil, fmt.Errorf("unknown Probe %q", req.Probe)
	}

	httpResp, body, err := doGET(probeURL)
	// Always tear down the session after the probe (success or fail).
	runCancel()
	out := <-runDone

	resp := &Response{
		RealSessionID: realSessionID,
		BaseURL:       baseURL,
		ProbeURL:      probeURL,
		RunExitCode:   -1,
	}
	if out.result != nil {
		resp.RunExitCode = out.result.ExitCode
	} else if out.err != nil {
		resp.RunExitCode = 1
	}
	if out.err != nil {
		resp.RunErrText = out.err.Error()
	}
	if err != nil {
		// Probe transport failure — surface as error to Assert.
		return resp, err
	}
	resp.StatusCode = httpResp.StatusCode
	resp.ContentType = httpResp.Header.Get("Content-Type")
	resp.Body = body
	resp.BodyString = string(body)

	if req.Probe == ProbeV1Session && len(body) > 0 && strings.Contains(strings.ToLower(resp.ContentType), "json") {
		parseSessionJSON(resp, body)
	} else if req.Probe == ProbeV1Session && len(body) > 0 && resp.StatusCode == http.StatusOK {
		// Content-Type may be missing in early implementations; still try JSON.
		parseSessionJSON(resp, body)
	}

	// Harness always returns (resp, nil) for HTTP-level outcomes so Assert owns
	// status-code expectations (including 404). Transport errors already returned.
	return resp, nil
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

func postHello(baseURL, version string, features []string, omitFeatures bool) error {
	payload := map[string]any{"version": version}
	if !omitFeatures {
		if features == nil {
			features = []string{}
		}
		payload["features"] = features
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	res, err := http.Post(baseURL+"/v1/hello", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	io.Copy(io.Discard, res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("hello status %d", res.StatusCode)
	}
	return nil
}

func postStatusRecording(baseURL string, entryCount, windowID int) error {
	payload := map[string]any{
		"state":       "recording",
		"entry_count": entryCount,
		"window_id":   windowID,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	res, err := http.Post(baseURL+"/v1/status", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	io.Copy(io.Discard, res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("status status %d", res.StatusCode)
	}
	return nil
}

func doGET(url string) (*http.Response, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res, nil, err
	}
	// Copy status/header we need; body already read.
	return res, body, nil
}

func parseSessionJSON(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.Raw = raw
	resp.SessionID, _ = raw["session_id"].(string)
	resp.Phase, _ = raw["phase"].(string)
	resp.Hint, _ = raw["hint"].(string)

	if ext, ok := raw["extension"].(map[string]any); ok {
		resp.ExtensionConnected, _ = ext["connected"].(bool)
		resp.ExtensionVersion, _ = ext["version"].(string)
		resp.SupportsBrowserTrace, _ = ext["supports_browser_trace"].(bool)
		if feats, ok := ext["features"].([]any); ok {
			for _, f := range feats {
				if s, ok := f.(string); ok {
					resp.ExtensionFeatures = append(resp.ExtensionFeatures, s)
				}
			}
		}
	}
	if rec, ok := raw["recording"].(map[string]any); ok {
		resp.RecordingActive, _ = rec["active"].(bool)
		resp.EntryCount = jsonInt(rec["entry_count"])
		resp.WindowID = jsonInt(rec["window_id"])
	}
	if ready, ok := raw["ready"].(map[string]any); ok {
		resp.ReadyDeadlineMS = int64(jsonInt(ready["deadline_ms"]))
		resp.ReadyElapsedMS = int64(jsonInt(ready["elapsed_ms"]))
		resp.ReadyRemainingMS = int64(jsonInt(ready["remaining_ms"]))
	}
}

func jsonInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}
```
