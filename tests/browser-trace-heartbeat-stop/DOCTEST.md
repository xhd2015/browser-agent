# browser-trace — install re-run guidance + recording heartbeat stop

Exercises two product surfaces added by the install/update re-run guidance and
recording **heartbeat stop** requirement:

1. **Session page install copy** (`GET /go`) — after Load unpacked / Reload, the
   user must **close the Chrome window and re-run `browser-trace`** (no mid-session
   hot-reload of capture).
2. **Recording heartbeat** — once recording is established, liveness refreshes on
   `POST /v1/status` and `POST /v1/entries`. If silent longer than
   `HeartbeatTimeout` (product default **10s**, injectable for tests), treat as
   browser-gone:
   - Save `recording.har` from the last `/v1/entries` snapshot (may be empty)
   - `meta.json`: `stop_reason=heartbeat_lost`, `partial=true`, entry_count
   - **Exit 0** with **stderr warning** (unusual but acceptable)
   - Stdout still ends with session dir path + `\n`

This tree is **separate** from:

- `./tests/browser-trace/` — sealed lifecycle / ready / complete timeouts
- `./tests/browser-trace-install-panel/` — always-visible install panel expand/collapse
- `./tests/browser-trace-exclude-clear-preview/` — live entries push + preview (not stop)
- `./tests/browser-trace-logging/` — ready-phase heartbeat logging (≠ recording heartbeat)

**No real Chrome**. HTTP and session leaves drive `browsertrace.Run` with
`NoOpenChrome` and an in-process **Mock Extension**. Ready-phase timeout semantics
are unchanged and **out of scope**.

Product defaults relevant here:

| Setting | Default |
|---------|---------|
| HeartbeatTimeout (recording) | **10s** (tests inject e.g. 200ms) |
| heartbeat_lost exit code | **0** |
| heartbeat_lost stdout | session path + `\n` |
| ReadyTimeout / CompleteTimeout | product 30s (tests shorten) |

## Version

0.0.2

# DSN (Domain Specific Notion)

**User** runs **browser-trace** to record network traffic. They may install or
reload the extension from the **Session Page** (`GET /go`). After Load unpacked
or Reload, capture does **not** hot-reload mid-session — the user must **close
the Chrome window** and **re-run `browser-trace`**.

**Control Server** owns one live **Session**:

- `lastHeartbeatAt` — refreshed on `POST /v1/status` and `POST /v1/entries`
  while recording
- `previewEntries` — last entries snapshot (source for partial HAR)
- `HeartbeatTimeout` from **Config** (default 10s)

**Recording Heartbeat Watcher** (after ready → recording):

- If no status/entries for longer than `HeartbeatTimeout` → **heartbeat_lost**
- Writes `recording.har` from `previewEntries` (empty array OK)
- Writes `meta.json` with `stop_reason=heartbeat_lost`, `partial=true`
- Emits stderr **warning** (tokens: warning, heartbeat, unusual/acceptable,
  snapshot/saved/closed)
- Exit **0**; stdout session path + trailing newline

**Normal complete** still works: continuous status/entries then
`POST /v1/complete` → exit 0, no heartbeat warning required.

**Install re-run guidance** on `/go` HTML:

- Server-rendered copy: close (this) browser/Chrome window; re-run `browser-trace`
  after install / Load unpacked / Reload
- Optional stable marker: `data-install-rerun-guidance`
- One-shot client banner after first healthy extension state is product UX;
  server-side copy / data attribute is enough for this tree (no DOM automation)

**Mock Extension** (tests):

- hello → start → status `recording` → optional POST entries → silence **or**
  continuous status/entries then complete

**Test Client** never opens Chrome; injects short `HeartbeatTimeout` on
heartbeat_lost leaves.

## Decision Tree

```
browser-trace-heartbeat-stop
├── go-install-rerun/                          [GET /go install re-run copy]
│   └── copy-present/                            close window + re-run browser-trace
└── recording-stop/                            [full session lifecycle]
    ├── heartbeat-lost/                        [silence > HeartbeatTimeout]
    │   ├── with-snapshot/                       POST entries (N URLs) then silence
    │   └── empty-snapshot/                      never POST entries; silence
    └── complete-ok/                           [normal POST /v1/complete]
        └── continuous-then-complete/            continuous status/entries + complete
```

### Parameter significance (high → low)

1. **Surface** — `/go` install re-run HTML vs full recording session stop path
   (different `Run` modes / outcomes).
2. **Stop path** (recording) — heartbeat_lost vs normal complete (exit semantics
   and artifacts differ).
3. **Snapshot presence** (heartbeat_lost) — non-empty last `/v1/entries` snapshot
   vs never posted (empty/minimal HAR); both still exit 0 + warning.

## Test Index

| Leaf | Scenario (requirement #) |
|------|--------------------------|
| `go-install-rerun/copy-present` | (#1) GET `/go` mentions close window + re-run `browser-trace` |
| `recording-stop/heartbeat-lost/with-snapshot` | (#3) POST entries with N URLs then silence → exit 0, warning, HAR URLs, partial meta |
| `recording-stop/heartbeat-lost/empty-snapshot` | (#4) Silence with no entries → exit 0, warning, empty HAR, partial meta |
| `recording-stop/complete-ok/continuous-then-complete` | (#2, #5) Continuous status/entries + complete → exit 0, normal success, no HB warning |

## How to Run

```sh
cd tests/browser-trace-heartbeat-stop
doctest vet .
doctest test -v .
# or from repo root:
doctest vet ./tests/browser-trace-heartbeat-stop
doctest test ./tests/browser-trace-heartbeat-stop
# regression (must stay green):
doctest test ./tests/browser-trace
doctest test ./tests/browser-trace-install-panel
```

Requires package `github.com/xhd2015/browser-agent/browsertrace` with
recording heartbeat + install re-run copy (TDD red until implementer lands them).

### Implementer contract

#### Config

```go
// HeartbeatTimeout is how long the recording phase may go without
// POST /v1/status or POST /v1/entries before heartbeat_lost.
// Zero → DefaultHeartbeatTimeout (10s). Injectable for tests (e.g. 200ms).
HeartbeatTimeout time.Duration
```

Default constant (suggested name): `DefaultHeartbeatTimeout = 10 * time.Second`.

Distinct from **ready-phase** `ReadyHeartbeat` (progress logging while waiting
for hello/recording).

#### Heartbeat refresh

While status is `recording` (and not yet complete), update `lastHeartbeatAt` on:

- successful `POST /v1/status`
- successful `POST /v1/entries`

#### heartbeat_lost outcome

When silence exceeds `HeartbeatTimeout` after recording established:

1. Save `{sessionDir}/recording.har` from last `previewEntries` snapshot
   (empty entries → valid HAR with `log.entries: []`).
2. Write `{sessionDir}/meta.json` with at least:
   - `stop_reason`: `"heartbeat_lost"` (or equivalent containing `heartbeat`)
   - `partial`: `true`
   - `entry_count`: count from snapshot (0 when empty)
3. Stderr includes a **warning** (case-insensitive token match):
   - `warning`
   - `heartbeat`
   - `unusual` **or** `acceptable`
   - and ideally snapshot/saved/closed language
4. **Exit code 0**; stdout ends with session path + `\n` (same success-like contract).

#### Normal complete

Unchanged: `POST /v1/complete` → exit 0, full HAR from complete payload, no
requirement to emit heartbeat warning.

#### `/go` install re-run copy

Server-rendered HTML for the session page must include guidance like:

- close this browser / Chrome window
- re-run / run `browser-trace` again
- after install / Load unpacked / Reload

Optional stable marker for tests: `data-install-rerun-guidance` on a panel or
paragraph. Client one-shot banner is optional and not multi-polled here.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browsertrace"
)

// Mode values — surface under test (set by grouping SETUP).
const (
	ModeGoHTML   = "go"
	ModeSession  = "session"
)

// ExtensionScript values for ModeSession (set by leaf SETUP).
const (
	// ExtNone — no mock extension.
	ExtNone = "none"
	// ExtSilenceWithSnapshot — hello, start, recording status, POST entries, then silence.
	ExtSilenceWithSnapshot = "silence-with-snapshot"
	// ExtSilenceEmpty — hello, start, recording status, then silence (no entries POST).
	ExtSilenceEmpty = "silence-empty"
	// ExtContinuousComplete — hello, start, keep status/entries, then POST complete.
	ExtContinuousComplete = "continuous-complete"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects which operation Run executes: ModeGoHTML or ModeSession.
	Mode string

	// Addr is host:port for Control Server. Empty → free loopback port in Run.
	Addr string
	// BaseDir is session parent directory (temp per leaf).
	BaseDir string
	// ReadyTimeout / CompleteTimeout override product defaults for fast tests.
	ReadyTimeout    time.Duration
	CompleteTimeout time.Duration
	// HeartbeatTimeout injects recording heartbeat timeout (0 → product default in package).
	// Heartbeat-lost leaves set a short value (e.g. 200ms).
	HeartbeatTimeout time.Duration
	// NoOpenChrome skips launching Chrome (always true in this tree).
	NoOpenChrome bool
	// SessionSuffix optional fixed session id / dir suffix when supported.
	SessionSuffix string

	// ExtensionScript selects mock extension behavior (ModeSession).
	ExtensionScript string
	// SnapshotURLs are request URLs embedded in POST /v1/entries for with-snapshot leaves.
	SnapshotURLs []string
	// MockStopReason on complete path (default "extension").
	MockStopReason string
	// MockWindowID reported in status/complete (default 42).
	MockWindowID int
	// ContinuousTicks is how many status/entries heartbeats before complete
	// (ExtContinuousComplete). Default 3.
	ContinuousTicks int
}

// Response is collected after the harness Run returns.
type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
	// ErrText is the string form of the error returned by browsertrace.Run, if any.
	ErrText string

	SessionDir string
	MetaPath   string
	HARPath    string
	MetaJSON   []byte
	HARJSON    []byte

	// HTTP probe (ModeGoHTML)
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string
	ProbeURL    string
	BaseURL     string

	// Mock observations (ModeSession)
	HelloOK           bool
	StatusRecording   bool
	EntriesPosted     bool
	CompletePosted    bool
	MockReceivedStart bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeGoHTML:
		return runGoHTML(t, req)
	case ModeSession:
		return runSession(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runGoHTML(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	req.NoOpenChrome = true
	if req.ReadyTimeout <= 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout <= 0 {
		req.CompleteTimeout = 2 * time.Second
	}
	if req.SessionSuffix == "" {
		req.SessionSuffix = "go-rerun-test"
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

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()

	var stdout, stderr bytes.Buffer
	cfg := browsertrace.Config{
		Addr:             addr,
		BaseDir:          req.BaseDir,
		ReadyTimeout:     req.ReadyTimeout,
		CompleteTimeout:  req.CompleteTimeout,
		HeartbeatTimeout: req.HeartbeatTimeout,
		NoOpenChrome:     true,
		SessionSuffix:    req.SessionSuffix,
		Stdout:           &stdout,
		Stderr:           &stderr,
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

	probeURL := baseURL + "/go?session=" + req.SessionSuffix
	httpResp, body, err := doGET(probeURL)
	runCancel()
	out := <-runDone

	resp := &Response{
		BaseURL:    baseURL,
		ProbeURL:   probeURL,
		Body:       body,
		BodyString: string(body),
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		ExitCode:   -1,
	}
	if err != nil {
		return resp, fmt.Errorf("GET /go: %w", err)
	}
	if httpResp != nil {
		resp.StatusCode = httpResp.StatusCode
		resp.ContentType = httpResp.Header.Get("Content-Type")
	}
	if out.err != nil {
		resp.ErrText = out.err.Error()
	}
	if out.result != nil {
		resp.ExitCode = out.result.ExitCode
		resp.SessionDir = out.result.SessionDir
	}
	return resp, nil
}

func runSession(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	req.NoOpenChrome = true
	if req.ReadyTimeout <= 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout <= 0 {
		req.CompleteTimeout = 5 * time.Second
	}
	if req.MockWindowID == 0 {
		req.MockWindowID = 42
	}
	if req.ContinuousTicks <= 0 {
		req.ContinuousTicks = 3
	}
	if req.ExtensionScript == "" {
		req.ExtensionScript = ExtNone
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
	resp := &Response{ExitCode: -1}
	var recording atomic.Bool

	sessionID := req.SessionSuffix
	if sessionID == "" {
		// Discover after server is up via first health+session dir; mock uses
		// whatever SessionSuffix product assigns. Prefer fixed suffix when set.
		sessionID = "hb-session"
		req.SessionSuffix = sessionID
	}

	var mock *mockExtension
	if req.ExtensionScript != ExtNone {
		mock = &mockExtension{
			baseURL:         baseURL,
			script:          req.ExtensionScript,
			windowID:        req.MockWindowID,
			sessionID:       sessionID,
			snapshotURLs:    append([]string(nil), req.SnapshotURLs...),
			stopReason:      req.MockStopReason,
			continuousTicks: req.ContinuousTicks,
			onRecording: func() {
				recording.Store(true)
				resp.StatusRecording = true
			},
		}
	}

	// Safety deadline so silence leaves fail fast if HeartbeatTimeout is not
	// implemented yet (TDD red) instead of hanging until the process is killed.
	maxWait := req.ReadyTimeout + req.CompleteTimeout + 3*time.Second
	if req.HeartbeatTimeout > 0 {
		// Allow several heartbeat periods + settle for save/exit.
		if hb := req.HeartbeatTimeout*8 + 2*time.Second; req.ReadyTimeout+hb > maxWait {
			maxWait = req.ReadyTimeout + hb
		}
	} else {
		// Continuous-complete uses product default heartbeat; still bound the test.
		maxWait = req.ReadyTimeout + 8*time.Second
	}
	runCtx, runCancel := context.WithTimeout(context.Background(), maxWait)
	defer runCancel()
	mockCtx, mockCancel := context.WithCancel(context.Background())
	defer mockCancel()

	var mockWG sync.WaitGroup
	if mock != nil {
		mockWG.Add(1)
		go func() {
			defer mockWG.Done()
			deadline := time.Now().Add(maxWait)
			for time.Now().Before(deadline) {
				if err := waitHealth(baseURL, 200*time.Millisecond); err == nil {
					mock.run(mockCtx)
					return
				}
				select {
				case <-mockCtx.Done():
					return
				case <-time.After(20 * time.Millisecond):
				}
			}
		}()
	}

	var stdout, stderr bytes.Buffer
	cfg := browsertrace.Config{
		Addr:             addr,
		BaseDir:          req.BaseDir,
		ReadyTimeout:     req.ReadyTimeout,
		CompleteTimeout:  req.CompleteTimeout,
		HeartbeatTimeout: req.HeartbeatTimeout,
		NoOpenChrome:     true,
		SessionSuffix:    req.SessionSuffix,
		Stdout:           &stdout,
		Stderr:           &stderr,
	}

	result, runErr := browsertrace.Run(runCtx, cfg)
	runCancel()
	mockCancel()
	mockWG.Wait()

	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	if runErr != nil {
		resp.ErrText = runErr.Error()
	}
	if result != nil {
		resp.ExitCode = result.ExitCode
		if result.SessionDir != "" {
			resp.SessionDir = result.SessionDir
		}
		if result.Stdout != "" && resp.Stdout == "" {
			resp.Stdout = result.Stdout
		}
		if result.Stderr != "" && resp.Stderr == "" {
			resp.Stderr = result.Stderr
		}
	} else if runErr != nil {
		resp.ExitCode = 1
	}

	if resp.SessionDir == "" {
		resp.SessionDir = findLatestSessionDir(req.BaseDir)
	}
	if resp.SessionDir != "" {
		resp.MetaPath = filepath.Join(resp.SessionDir, "meta.json")
		resp.HARPath = filepath.Join(resp.SessionDir, "recording.har")
		if b, err := os.ReadFile(resp.MetaPath); err == nil {
			resp.MetaJSON = b
		}
		if b, err := os.ReadFile(resp.HARPath); err == nil {
			resp.HARJSON = b
		}
	}
	if mock != nil {
		resp.MockReceivedStart = mock.gotStart
		resp.HelloOK = mock.didHello
		resp.CompletePosted = mock.didComplete
		resp.EntriesPosted = mock.didEntries
		if mock.didRecordingStatus {
			resp.StatusRecording = true
		}
	}

	// Always return (resp, nil) so Assert inspects ExitCode / stderr / artifacts.
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
	// Clone header-bearing response for callers (body already drained).
	cloned := &http.Response{
		StatusCode: res.StatusCode,
		Header:     res.Header.Clone(),
	}
	return cloned, body, nil
}

func findLatestSessionDir(base string) string {
	entries, err := os.ReadDir(base)
	if err != nil {
		return ""
	}
	var best string
	var bestMod time.Time
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Skip extension extract dir if present.
		if e.Name() == "extension" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if best == "" || info.ModTime().After(bestMod) {
			best = filepath.Join(base, e.Name())
			bestMod = info.ModTime()
		}
	}
	return best
}

// mockExtension is an in-process client of the control protocol for this tree.
type mockExtension struct {
	baseURL         string
	script          string
	windowID        int
	sessionID       string
	snapshotURLs    []string
	stopReason      string
	continuousTicks int
	onRecording     func()

	gotStart           bool
	didHello           bool
	didRecordingStatus bool
	didEntries         bool
	didComplete        bool
}

func (m *mockExtension) run(ctx context.Context) {
	switch m.script {
	case ExtNone:
		return
	case ExtSilenceWithSnapshot, ExtSilenceEmpty, ExtContinuousComplete:
		m.postHello()
		if !m.waitCommand(ctx, "start") {
			return
		}
		m.gotStart = true
		m.postStatus("recording", 0)
		m.didRecordingStatus = true
		if m.onRecording != nil {
			m.onRecording()
		}

		switch m.script {
		case ExtSilenceWithSnapshot:
			urls := m.snapshotURLs
			if len(urls) == 0 {
				urls = []string{
					"https://api.example.com/v1/alpha",
					"https://cdn.example.com/app.js",
				}
			}
			if err := m.postEntries(urls); err == nil {
				m.didEntries = true
			}
			// One more status to refresh heartbeat with entry_count, then silence.
			m.postStatus("recording", len(urls))
			<-ctx.Done()
		case ExtSilenceEmpty:
			// No entries POST — server has empty preview snapshot.
			<-ctx.Done()
		case ExtContinuousComplete:
			n := m.continuousTicks
			if n <= 0 {
				n = 3
			}
			for i := 0; i < n; i++ {
				select {
				case <-ctx.Done():
					return
				case <-time.After(40 * time.Millisecond):
				}
				urls := m.snapshotURLs
				if len(urls) == 0 {
					urls = []string{"https://api.example.com/live"}
				}
				_ = m.postEntries(urls)
				m.didEntries = true
				m.postStatus("recording", len(urls))
			}
			if m.stopReason == "" {
				m.stopReason = "extension"
			}
			m.postComplete(urlsOrDefault(m.snapshotURLs))
			m.didComplete = true
		}
	default:
		// Unknown script: idle.
		<-ctx.Done()
	}
}

func urlsOrDefault(urls []string) []string {
	if len(urls) == 0 {
		return []string{"https://api.example.com/live"}
	}
	return urls
}

func (m *mockExtension) postHello() {
	body := `{"version":"1.2.0","features":["browser-trace"]}`
	res, err := http.Post(m.baseURL+"/v1/hello", "application/json", strings.NewReader(body))
	if err != nil {
		return
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		m.didHello = true
	}
}

func (m *mockExtension) postStatus(state string, entryCount int) {
	payload := map[string]any{
		"state":       state,
		"entry_count": entryCount,
		"window_id":   m.windowID,
	}
	b, _ := json.Marshal(payload)
	res, err := http.Post(m.baseURL+"/v1/status", "application/json", bytes.NewReader(b))
	if err != nil {
		return
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
}

func (m *mockExtension) postEntries(urls []string) error {
	entries := make([]map[string]any, 0, len(urls))
	for i, u := range urls {
		entries = append(entries, map[string]any{
			"startedDateTime": fmt.Sprintf("2026-07-11T12:00:%02d.000Z", i),
			"request": map[string]any{
				"method":      "GET",
				"url":         u,
				"httpVersion": "HTTP/1.1",
				"headers":     []any{},
			},
			"response": map[string]any{
				"status":      200,
				"statusText":  "OK",
				"httpVersion": "HTTP/1.1",
				"headers":     []any{},
				"content":     map[string]any{"size": 0, "mimeType": "text/plain"},
			},
			"time": 1,
		})
	}
	payload := map[string]any{
		"session_id": m.sessionID,
		"entries":    entries,
		"count":      len(entries),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/v1/entries", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("POST /v1/entries status %d", res.StatusCode)
	}
	return nil
}

func (m *mockExtension) postComplete(urls []string) {
	entries := make([]any, 0, len(urls))
	for i, u := range urls {
		entries = append(entries, map[string]any{
			"startedDateTime": fmt.Sprintf("2026-07-11T12:00:%02d.000Z", i),
			"request": map[string]any{
				"method": "GET",
				"url":    u,
			},
			"response": map[string]any{
				"status": 200,
			},
			"time": 1,
		})
	}
	har := map[string]any{
		"log": map[string]any{
			"version": "1.2",
			"creator": map[string]any{"name": "test-mock", "version": "1.0"},
			"entries": entries,
		},
	}
	payload := map[string]any{
		"har":         har,
		"stop_reason": m.stopReason,
		"window_id":   m.windowID,
		"stats":       map[string]any{"entry_count": len(entries)},
	}
	b, _ := json.Marshal(payload)
	res, err := http.Post(m.baseURL+"/v1/complete", "application/json", bytes.NewReader(b))
	if err != nil {
		return
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
}

func (m *mockExtension) waitCommand(ctx context.Context, want string) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+"/v1/commands?wait=2", nil)
		if err != nil {
			return false
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(30 * time.Millisecond):
				continue
			}
		}
		raw, _ := io.ReadAll(res.Body)
		res.Body.Close()
		var cmd struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(raw, &cmd)
		if cmd.Type == want {
			return true
		}
	}
}
```
