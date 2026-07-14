# browser-trace — auto-exclude control traffic, live entries push, server preview

Exercises two product surfaces for **live capture preview** and **control-server
auto-exclude**:

1. **Pure filter** `ShouldCaptureURL(url string) bool` — never capture requests to
   the product control hosts `http://127.0.0.1:43759/…` and
   `http://localhost:43759/…`.
2. **Control server HTTP** — mock extension **push** of in-memory HAR entries
   and the live HTML viewer:
   - `POST /v1/entries` — snapshot `{session_id, entries, count}` (including empty after clear)
   - `GET /v1/entries?session=` — last snapshot as JSON
   - `GET /preview?session=` — HTML live viewer (polls `/v1/entries`)

This tree is **separate** from:

- `./tests/browser-trace/` — sealed lifecycle / HAR complete / deadlines
- `./tests/browser-trace-session-page/` — `/v1/session` + `/go` status UI
- `./tests/browser-trace-install-panel/` — install panel expand/collapse
- `./tests/browser-trace-logging/` — ready-wait logging

**No real Chrome**, no extension process, no popup/DOM automation. HTTP leaves
drive `browsertrace.Run` with `NoOpenChrome` and an in-process client that
simulates the extension’s periodic entries push. Pure-filter leaves call the Go
helper only.

Product defaults relevant here:

| Setting | Default |
|---------|---------|
| control hosts (exclude) | `127.0.0.1:43759`, `localhost:43759` |
| listen address (product) | `127.0.0.1:43759` (tests use ephemeral free port) |
| entries push interval (product) | ~1s while recording (harness posts immediately) |
| clear | extension empties local entries then `POST` empty snapshot |
| preview preferred URL | `http://127.0.0.1:43759/preview?session=…` |

## Version

0.0.2

# DSN (Domain Specific Notion)

**User** records with **browser-trace**. The **Extension Agent** captures page
network traffic but must **not** record its own traffic to the **Control Server**
(hello, status, commands, entries push, preview). The pure filter
`ShouldCaptureURL` is the gate before inserting into in-memory `entries`.

While **recording**, the agent periodically **POSTs** the current entries
snapshot to the Control Server so a **live Preview** can show progress without
waiting for final HAR on stop. The popup **Clear captured** action empties local
entries (state stays `recording`) and immediately POSTs an **empty** snapshot so
the server preview resets. Preview opens the server `/preview?session=` page when
the control session is available (fallback to extension `preview.html` is
product-only and **out of scope** for this tree).

**Control Server** (per live session, in memory):

- `previewEntries []json` — last POST snapshot
- `previewUpdatedAt time` — last update timestamp
- `POST /v1/entries` — body `{ "session_id", "entries": […], "count": N }`
- `GET /v1/entries?session=` — `{ "entries", "count", "updated_at" }`
- `GET /preview?session=` — `text/html` viewer that polls `/v1/entries`
- Unknown session id → **HTTP 404** (JSON for entries API; HTML or JSON error for preview)

**Test Client** in this tree:

- Pure mode: call `browsertrace.ShouldCaptureURL` only
- HTTP mode: start control server via `browsertrace.Run` (`NoOpenChrome`, known
  `SessionSuffix`, temp `BaseDir`, free port), optionally POST entry snapshots
  (mock extension push / clear), then GET `/v1/entries` or `/preview`

## Decision Tree

```
browser-trace-exclude-clear-preview
├── should-capture-url/                          [pure ShouldCaptureURL]
│   ├── reject/                                    expect false (control traffic)
│   │   ├── control-ip/                              host 127.0.0.1:43759
│   │   │   ├── root/                                  http://127.0.0.1:43759/
│   │   │   └── with-path/                             …/v1/entries (path+query)
│   │   └── control-localhost/                       host localhost:43759
│   │       ├── root/                                  http://localhost:43759/
│   │       └── with-path/                             …/preview?session=x
│   └── allow/                                     expect true (not control)
│       ├── normal-https/                            https://api.example.com/…
│       └── other-loopback-port/                     http://127.0.0.1:8080/… (≠43759)
└── control-server-http/                         [HTTP mock extension push]
    ├── v1-entries/                                POST/GET /v1/entries JSON
    │   ├── session-known/
    │   │   ├── post-then-get/                       POST snapshot → GET count+URLs
    │   │   └── clear-empty/                         POST data then POST empty → GET 0
    │   └── session-missing/                         unknown session → 404 JSON
    └── preview/                                   GET /preview HTML
        ├── session-known/
        │   ├── with-entries/                        200 HTML has entry URL + poll marker
        │   └── empty-after-clear/                   200 HTML empty/cleared state
        └── session-missing/                         unknown session → 404 error
```

### Parameter significance (high → low)

1. **Surface** — pure `ShouldCaptureURL` vs live control-server HTTP (different
   `Run` modes / no shared server needed for pure).
2. **Capture decision** (pure) — reject control traffic vs allow normal traffic.
3. **Control host form** (reject) — product exclude must cover **both**
   `127.0.0.1:43759` and `localhost:43759`.
4. **Path shape** (reject) — root `/` vs resource path/query (all under control host).
5. **HTTP probe** — `/v1/entries` JSON API vs `/preview` HTML viewer.
6. **Session identity** — known live session vs unknown id (404).
7. **Snapshot sequence** (known session) — post-then-get vs clear-empty
   (empty POST after data); for preview, with-entries vs empty-after-clear.

## Test Index

| Leaf | Scenario (requirement #) |
|------|--------------------------|
| `should-capture-url/reject/control-ip/root` | (#1) `http://127.0.0.1:43759/` → false |
| `should-capture-url/reject/control-ip/with-path` | (#1) control IP path/query → false |
| `should-capture-url/reject/control-localhost/root` | (#1) `http://localhost:43759/` → false |
| `should-capture-url/reject/control-localhost/with-path` | (#1) localhost path → false |
| `should-capture-url/allow/normal-https` | (#2) normal HTTPS API URL → true |
| `should-capture-url/allow/other-loopback-port` | (#2 ext) other loopback port → true |
| `control-server-http/v1-entries/session-known/post-then-get` | (#3) POST then GET match count+URLs |
| `control-server-http/v1-entries/session-known/clear-empty` | (#4) POST empty after clear → GET count 0 |
| `control-server-http/v1-entries/session-missing` | (#6) unknown session entries → 404 JSON |
| `control-server-http/preview/session-known/with-entries` | (#5) preview HTML has URL + poll/entries marker |
| `control-server-http/preview/session-known/empty-after-clear` | (#5/#4) preview after empty push |
| `control-server-http/preview/session-missing` | (#6) unknown session preview → 404 |

## How to Run

```sh
cd tests/browser-trace-exclude-clear-preview
doctest vet .
doctest test -v .
# or from repo root:
doctest vet ./tests/browser-trace-exclude-clear-preview
doctest test ./tests/browser-trace-exclude-clear-preview
# regression (lifecycle + session surfaces must stay green):
doctest test ./tests/browser-trace
doctest test ./tests/browser-trace-session-page
```

Requires package `github.com/xhd2015/browser-agent/browsertrace`:

- Existing `Config` + `Run(ctx, Config) (*Result, error)` control server
- New pure helper: `ShouldCaptureURL(url string) bool`
- New routes: `POST/GET /v1/entries`, `GET /preview`
- Leaves fail to compile/run until implementer adds them (TDD red → green)

### Expected contracts (implementer)

#### Pure filter

```go
// ShouldCaptureURL reports whether a request URL should be stored in the
// capture buffer. Returns false for traffic to the product control hosts:
//   http://127.0.0.1:43759  and  http://localhost:43759
// (any path, query, or fragment). All other well-formed http(s) URLs return true.
func ShouldCaptureURL(url string) bool
```

Notes:

- Port **43759** is the product control port (`DefaultAddr`); exclusion is by
  host+port, not by the ephemeral bind address used in tests.
- Scheme in product examples is `http`; implementations may also treat
  `https://127.0.0.1:43759` as control if convenient, but **doctest asserts
  only the http cases listed in leaves**.
- Invalid/empty URL handling is implementer choice; not covered here.

#### `POST /v1/entries`

```http
POST /v1/entries
Content-Type: application/json

{
  "session_id": "<id>",
  "entries": [ { "request": { "url": "https://api.example.com/a", "method": "GET" }, … } ],
  "count": 1
}
```

- Known session → 2xx; server stores snapshot (replace, not merge required).
- Unknown session → **404** JSON indicating not found.
- Empty `entries` + `count: 0` is the **clear** snapshot (preview resets).

#### `GET /v1/entries?session=`

```json
{
  "entries": [ … ],
  "count": N,
  "updated_at": "<RFC3339 or unix; non-empty when a snapshot exists>"
}
```

- Known session with no POST yet: `count` 0 / empty entries is acceptable.
- Unknown session → **404** JSON not-found.

#### `GET /preview?session=`

- Known session → **200** `text/html` (or HTML content-type).
- Body should reference live data: either embed entry URL(s) and/or client JS
  that polls `/v1/entries` (or otherwise marks a live preview root).
- Empty snapshot: empty-state wording / zero-count marker / empty table OK.
- Unknown session → **404** (HTML or JSON error body indicating not found).

#### Harness notes

- HTTP leaves use free loopback port + temp `BaseDir`; probes run while
  `browsertrace.Run` is still in ready-wait; cancel after final probe.
- Fixture entries intentionally use **non-control** URLs only (exclude is
  client-side; server stores whatever is posted).

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

// Mode values — surface under test (set by grouping SETUP).
const (
	ModeShouldCapture = "should-capture"
	ModeHTTP          = "http"
)

// Probe targets for ModeHTTP (Request.Probe).
const (
	ProbeV1Entries = "v1-entries"
	ProbePreview   = "preview"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects pure filter vs control-server HTTP.
	// ModeShouldCapture or ModeHTTP.
	Mode string

	// --- pure ShouldCaptureURL ---
	// CaptureURL is the input to ShouldCaptureURL.
	CaptureURL string
	// WantCapture is the expected boolean result.
	WantCapture bool

	// --- HTTP control server ---
	// BaseDir is session parent directory (temp per leaf).
	BaseDir string
	// Addr is host:port. Empty → free loopback port in Run.
	Addr string
	// ReadyTimeout / CompleteTimeout override product defaults for fast tests.
	ReadyTimeout    time.Duration
	CompleteTimeout time.Duration
	// NoOpenChrome skips launching Chrome (always true for HTTP leaves).
	NoOpenChrome bool
	// SessionSuffix is the known session id (browsertrace uses suffix as id).
	SessionSuffix string

	// Probe selects final GET: ProbeV1Entries or ProbePreview.
	Probe string

	// ForceUnknownSession, when true, uses a non-live session id on the final
	// probe (and on any POST when PostUsesProbeSession is true).
	ForceUnknownSession bool
	// SessionIDForProbe overrides the session id used on probe (and optional POST).
	// Default unknown id is "does-not-exist".
	SessionIDForProbe string
	// PostUsesProbeSession, when true, POSTs use the same session id as the probe
	// (for missing-session POST attempts). Default false: POST uses live id.
	PostUsesProbeSession bool

	// StageEntries, when non-nil, is POSTed to /v1/entries before optional clear
	// and before the final GET probe. nil means skip initial POST.
	// Use empty non-nil slice for an intentional empty first post.
	StageEntries []map[string]any
	// DoStagePost, when true, POST StageEntries (even if empty) before probe.
	DoStagePost bool
	// DoClearAfterStage, when true, after Stage POST, POST empty entries
	// (clear captured → preview reset).
	DoClearAfterStage bool
}

// Response holds pure-filter and/or HTTP probe results.
type Response struct {
	// Pure filter
	CaptureResult bool
	CaptureCalled bool

	// HTTP probe
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string

	// Parsed GET /v1/entries JSON (when applicable).
	EntriesCount   int
	Entries        []map[string]any
	UpdatedAt      string
	EntriesRaw     map[string]any
	// EntryURLs collected from parsed entries request.url when present.
	EntryURLs []string

	// POST outcomes (last non-clear stage and/or clear), informational.
	StagePostStatus int
	ClearPostStatus int

	RealSessionID string
	BaseURL       string
	ProbeURL      string
	RunExitCode   int
	RunErrText    string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeShouldCapture:
		return runShouldCapture(t, req)
	case ModeHTTP:
		return runHTTP(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runShouldCapture(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CaptureURL == "" {
		t.Fatal("CaptureURL must be set for ModeShouldCapture")
	}
	got := browsertrace.ShouldCaptureURL(req.CaptureURL)
	return &Response{
		CaptureResult: got,
		CaptureCalled: true,
	}, nil
}

func runHTTP(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	if req.SessionSuffix == "" {
		t.Fatal("SessionSuffix must be set by Setup (known session id)")
	}
	if req.Probe == "" {
		t.Fatal("Probe must be set by grouping Setup")
	}
	req.NoOpenChrome = true
	if req.ReadyTimeout <= 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout <= 0 {
		req.CompleteTimeout = 2 * time.Second
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

	if err := waitHealth(baseURL, req.ReadyTimeout); err != nil {
		runCancel()
		<-runDone
		return nil, fmt.Errorf("control server never became healthy at %s: %w", baseURL, err)
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

	postSessionID := realSessionID
	if req.PostUsesProbeSession {
		postSessionID = probeSessionID
	}

	resp := &Response{
		RealSessionID: realSessionID,
		BaseURL:       baseURL,
		RunExitCode:   -1,
	}

	// Optional mock extension push (and clear).
	if req.DoStagePost {
		status, err := postEntries(baseURL, postSessionID, req.StageEntries)
		resp.StagePostStatus = status
		if err != nil && !req.ForceUnknownSession {
			// Transport failure on known session is fatal to harness.
			runCancel()
			<-runDone
			return resp, fmt.Errorf("POST /v1/entries (stage): %w", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	if req.DoClearAfterStage {
		status, err := postEntries(baseURL, postSessionID, []map[string]any{})
		resp.ClearPostStatus = status
		if err != nil && !req.ForceUnknownSession {
			runCancel()
			<-runDone
			return resp, fmt.Errorf("POST /v1/entries (clear): %w", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	var probeURL string
	switch req.Probe {
	case ProbeV1Entries:
		probeURL = baseURL + "/v1/entries?session=" + probeSessionID
	case ProbePreview:
		probeURL = baseURL + "/preview?session=" + probeSessionID
	default:
		runCancel()
		<-runDone
		return nil, fmt.Errorf("unknown Probe %q", req.Probe)
	}
	resp.ProbeURL = probeURL

	httpResp, body, err := doGET(probeURL)
	runCancel()
	out := <-runDone

	if out.result != nil {
		resp.RunExitCode = out.result.ExitCode
	} else if out.err != nil {
		resp.RunExitCode = 1
	}
	if out.err != nil {
		resp.RunErrText = out.err.Error()
	}
	if err != nil {
		return resp, err
	}
	resp.StatusCode = httpResp.StatusCode
	resp.ContentType = httpResp.Header.Get("Content-Type")
	resp.Body = body
	resp.BodyString = string(body)

	if req.Probe == ProbeV1Entries && len(body) > 0 {
		// Parse when JSON-ish (success or error body).
		parseEntriesJSON(resp, body)
	}

	// HTTP outcomes (incl. 404) are owned by Assert — return nil error.
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

// postEntries POSTs a full snapshot. entries may be empty (clear).
// Returns HTTP status and error (non-2xx is still returned with status, nil err
// for Assert-owned cases; transport errors return err).
func postEntries(baseURL, sessionID string, entries []map[string]any) (int, error) {
	if entries == nil {
		entries = []map[string]any{}
	}
	payload := map[string]any{
		"session_id": sessionID,
		"entries":    entries,
		"count":      len(entries),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/entries", bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	io.Copy(io.Discard, res.Body)
	return res.StatusCode, nil
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
	return res, body, nil
}

func parseEntriesJSON(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.EntriesRaw = raw
	resp.EntriesCount = jsonInt(raw["count"])
	if s, ok := raw["updated_at"].(string); ok {
		resp.UpdatedAt = s
	} else if n, ok := raw["updated_at"].(float64); ok {
		resp.UpdatedAt = fmt.Sprintf("%v", n)
	}
	if arr, ok := raw["entries"].([]any); ok {
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			resp.Entries = append(resp.Entries, m)
			if u := entryURL(m); u != "" {
				resp.EntryURLs = append(resp.EntryURLs, u)
			}
		}
		// If count omitted, derive from array length.
		if raw["count"] == nil {
			resp.EntriesCount = len(resp.Entries)
		}
	}
}

func entryURL(entry map[string]any) string {
	if req, ok := entry["request"].(map[string]any); ok {
		if u, ok := req["url"].(string); ok {
			return u
		}
	}
	if u, ok := entry["url"].(string); ok {
		return u
	}
	return ""
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

// Silence unused import guard.
var _ = strings.TrimSpace
```
