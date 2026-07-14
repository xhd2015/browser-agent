# browser-trace CLI — control server, deadlines, HAR save

Exercises the **`browser-trace`** CLI / `browsertrace` Go package: fixed-address
local control server, extension wire protocol, ready/complete deadlines, multi-tab
HAR + meta save, and user-facing stdout. Tests drive an **in-process mock
extension** over HTTP; **no real Chrome** and no real extension process.

Product defaults (overridable in tests):

| Flag / setting | Default |
|----------------|---------|
| listen address | `127.0.0.1:43759` |
| base dir | `~/.tmp/browser-trace` |
| ready timeout | `30s` |
| complete timeout | `30s` |
| open Chrome | yes (new window to session page) |

In this tree: free ephemeral `Addr`, temp `BaseDir`, short timeouts on failure
paths, and `NoOpenChrome=true` always.

## Version

0.0.2

# DSN (Domain Specific Notion)

**User** runs **browser-trace** to record network traffic from one new Chrome
window. The CLI owns a **Control Server** bound to a fixed loopback address.

**Control Server** exposes a small HTTP protocol for the extension agent:

- `GET /v1/health` — liveness
- `GET /go?session=…` — session page (Chrome target URL)
- `POST /v1/hello` — extension announces presence (version)
- `GET /v1/commands?wait=…` — long-poll next command (`start` / `stop` / noop)
- `POST /v1/status` — heartbeat (state, entry_count, window_id, tab attach)
- `POST /v1/complete` — final HAR JSON + stop_reason + stats

**Session** progresses:
`waiting_extension` → `recording` → `stopping` → `saved` | `failed`.

**Chrome Launcher** opens Google Chrome default profile in a **new window** to
the session page. In tests the launcher is a no-op (`NoOpenChrome`).

**Extension Agent** (product: Chrome-Ext-Capture-API) long-polls commands, on
`start` pins the session window and attaches Network capture to **all tabs in
that window**, merges multi-tab entries, and on Stop (popup or server `stop`)
POSTs `/v1/complete`. In tests a **Mock Extension** plays the same HTTP role
inside `Run`.

**Storage** writes under `{BaseDir}/YYYY-MM-DD-HH-MM-SS-<suffix>/`:

- `meta.json` — session id, times, stop_reason, entry_count, window_id, errors
- `recording.har` — HAR 1.2, multi-tab entries, sorted by `startedDateTime`
  (atomic write via `*.tmp` + rename where practical)

**Deadlines**:

- **Ready**: after listen, wait up to ready-timeout for hello **and** status
  `recording`; else fail with extension install/enable/host_permissions hint.
- **Complete**: after stop, wait up to complete-timeout for `/v1/complete`;
  else fail without overwriting a good final HAR.

**Port policy**: bind failure (address in use) is a **hard error** — no fallback
port. **Stdout** for user-facing success ends with a trailing newline `\n`.

## Decision Tree

```
browser-trace session lifecycle
├── listen-fail/                         [bind fails]
│   └── address-in-use                     port busy → exit ≠ 0, message mentions in use
└── listen-ok/                           [bind succeeds]
    ├── ready-fail/                      [never reaches recording within ready-timeout]
    │   ├── no-hello                       mock silent; no POST /v1/hello
    │   └── hello-no-recording             hello only (or non-recording status)
    └── ready-ok/                        [hello + status recording before ready-timeout]
        ├── complete-fail/               [stop issued; complete never arrives]
        │   └── no-complete-after-stop     exit ≠ 0; no corrupt final HAR
        └── complete-ok/                 [POST /v1/complete within complete-timeout]
            ├── extension-stop/          [stop_reason from extension complete]
            │   └── multi-tab-save-and-stdout
            │                              exit 0; merged HAR + meta; dir name; stdout \n
            └── cli-stop/                [CLI cancel/signal queues stop]
                └── signal-then-complete   mock sees stop; files written; exit 0
```

### Parameter significance (high → low)

1. **Listen outcome** — cannot run any session if bind fails (hard fail, no port fallback).
2. **Ready outcome** — extension must hello and reach `recording` within ready-timeout.
3. **Complete outcome** — after stop, final HAR must arrive within complete-timeout.
4. **Stop initiator** — extension complete vs CLI signal/cancel (same save contract).
5. **Artifact details** — multi-tab merge, session dir pattern, stdout trailing newline
   (asserted on the primary success leaf).

## Test Index

| Leaf | Scenario (requirement #) |
|------|--------------------------|
| `listen-fail/address-in-use` | (#1) Address already in use → hard fail |
| `listen-ok/ready-fail/no-hello` | (#2) No extension hello within ready timeout |
| `listen-ok/ready-fail/hello-no-recording` | (#3) Hello but never reaches recording |
| `listen-ok/ready-ok/complete-fail/no-complete-after-stop` | (#6) Stop then complete timeout |
| `listen-ok/ready-ok/complete-ok/extension-stop/multi-tab-save-and-stdout` | (#4, #7, #8) Multi-tab HAR + meta, dir naming, stdout `\n` |
| `listen-ok/ready-ok/complete-ok/cli-stop/signal-then-complete` | (#5) CLI stop while mock connected |

## How to Run

```sh
cd tests/browser-trace
doctest vet .
doctest test -v .
# or from repo root:
doctest vet ./tests/browser-trace
doctest test ./tests/browser-trace
```

Requires implementer package `github.com/xhd2015/browser-agent/browsertrace`
(and optional `cmd/browser-trace`). Tests intentionally fail to compile/run until
that package exists — TDD red phase.

### Expected `browsertrace` API (implementer contract)

Prose contract (not a harness code block): package
`github.com/xhd2015/browser-agent/browsertrace` must export:

- `Config` with fields: `Addr`, `BaseDir`, `ReadyTimeout`, `CompleteTimeout`,
  `NoOpenChrome`, `SessionSuffix`, `Stdout io.Writer`, `Stderr io.Writer`
- `Result` with fields: `ExitCode int`, `SessionDir string`, optional `Stdout`/`Stderr`
- `func Run(ctx context.Context, cfg Config) (*Result, error)` — bind Addr, run one
  session until saved|failed|timeouts; **ctx cancel** = CLI SIGINT/SIGTERM (queue
  `stop`, wait `CompleteTimeout` for HAR)

Wire protocol paths: `/v1/health`, `/go`, `/v1/hello`, `/v1/commands`, `/v1/status`, `/v1/complete`.

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

// ExtensionScript values (set by SETUP chain).
const (
	ExtNone               = "none"
	ExtHelloOnly          = "hello-only"
	ExtHelloNoRecording   = "hello-no-recording"
	ExtRecordAndComplete  = "record-and-complete"
	ExtRecordNoComplete   = "record-no-complete"
)

// StopMode values.
const (
	StopNone      = "none"
	StopExtension = "extension"
	StopCLI       = "cli-signal"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Addr is host:port for Control Server. Empty → free loopback port.
	Addr string
	// BaseDir is session parent directory (temp per leaf).
	BaseDir string
	// ReadyTimeout / CompleteTimeout override product 30s defaults in tests.
	ReadyTimeout    time.Duration
	CompleteTimeout time.Duration
	// NoOpenChrome skips launching Chrome (always true in this tree).
	NoOpenChrome bool

	// OccupyAddr, when true, binds a real listener on Addr before Run starts
	// browser-trace (port-conflict scenario).
	OccupyAddr bool

	// ExtensionScript selects mock extension behavior after the server is up.
	ExtensionScript string
	// StopMode selects how stop is initiated once recording is observed.
	// For ExtRecordAndComplete + StopExtension, mock POSTs complete itself.
	// For ExtRecordAndComplete + StopCLI, Run cancels ctx; mock completes on stop cmd.
	// For ExtRecordNoComplete, mock reaches recording then never POSTs complete
	// (optionally still answers stop via commands poll).
	StopMode string

	// SessionSuffix optional fixed suffix for out-dir naming (if package supports).
	SessionSuffix string

	// MockHAR is the HAR object (or full document) the mock POSTs on complete.
	// When empty, Run builds a multi-tab sample HAR.
	MockHAR json.RawMessage
	// MockStopReason is stop_reason field on complete (default "extension" or "cli").
	MockStopReason string
	// MockWindowID reported in status/complete (default 42).
	MockWindowID int
}

// Response is collected after browsertrace.Run returns.
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

	// Mock observations
	MockReceivedStart bool
	MockReceivedStop  bool
	HelloOK           bool
	StatusRecording   bool
	CompletePosted    bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	if req.NoOpenChrome == false {
		// Safety: never open real Chrome from this tree.
		req.NoOpenChrome = true
	}
	if req.ReadyTimeout <= 0 {
		req.ReadyTimeout = 30 * time.Second
	}
	if req.CompleteTimeout <= 0 {
		req.CompleteTimeout = 30 * time.Second
	}
	if req.MockWindowID == 0 {
		req.MockWindowID = 42
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

	var occupied net.Listener
	if req.OccupyAddr {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("test occupy listen %s: %w", addr, err)
		}
		occupied = ln
		defer occupied.Close()
	}

	baseURL := "http://" + addr
	resp := &Response{ExitCode: -1}
	var recording atomic.Bool

	var mock *mockExtension
	if req.ExtensionScript != ExtNone && req.ExtensionScript != "" {
		mock = &mockExtension{
			baseURL:    baseURL,
			script:     req.ExtensionScript,
			stopMode:   req.StopMode,
			windowID:   req.MockWindowID,
			har:        req.MockHAR,
			stopReason: req.MockStopReason,
			onRecording: func() {
				recording.Store(true)
				resp.StatusRecording = true
			},
		}
	}

	// runCtx is cancelled to simulate CLI SIGINT/SIGTERM for browsertrace.Run.
	// mockCtx is independent: a real extension does not share the CLI process
	// context, so after Ctrl-C it must still long-poll stop and POST complete.
	// Sharing one ctx made StopCLI impossible (cancel aborted mock HTTP).
	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()
	mockCtx, mockCancel := context.WithCancel(context.Background())
	defer mockCancel()

	// Start mock after a short delay so listen can fail fast on conflict.
	var mockWG sync.WaitGroup
	if mock != nil {
		mockWG.Add(1)
		go func() {
			defer mockWG.Done()
			// Wait for health (or give up with ready timeout window).
			deadline := time.Now().Add(req.ReadyTimeout + 2*time.Second)
			for time.Now().Before(deadline) {
				if occupied != nil {
					// Server should never come up; still exit when mockCtx done.
					select {
					case <-mockCtx.Done():
						return
					case <-time.After(50 * time.Millisecond):
						continue
					}
				}
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

	// For CLI stop: after mock reports recording, cancel runCtx (not mockCtx).
	if mock != nil && req.StopMode == StopCLI {
		go func() {
			deadline := time.Now().Add(req.ReadyTimeout + req.CompleteTimeout + 5*time.Second)
			for time.Now().Before(deadline) {
				if recording.Load() {
					// Small settle so status has been processed by server.
					time.Sleep(50 * time.Millisecond)
					runCancel()
					return
				}
				select {
				case <-runCtx.Done():
					return
				case <-time.After(20 * time.Millisecond):
				}
			}
		}()
	}

	var stdout, stderr bytes.Buffer
	cfg := browsertrace.Config{
		Addr:            addr,
		BaseDir:         req.BaseDir,
		ReadyTimeout:    req.ReadyTimeout,
		CompleteTimeout: req.CompleteTimeout,
		NoOpenChrome:    req.NoOpenChrome,
		SessionSuffix:   req.SessionSuffix,
		Stdout:          &stdout,
		Stderr:          &stderr,
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
		// Bind / early failures may only return error.
		resp.ExitCode = 1
	}

	// Discover session dir under BaseDir if not provided.
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
		resp.MockReceivedStop = mock.gotStop
		resp.HelloOK = mock.didHello
		resp.CompletePosted = mock.didComplete
		if mock.didRecordingStatus {
			resp.StatusRecording = true
		}
	}

	// Run always returns (resp, nil) so Assert can inspect ExitCode/ErrText;
	// transport failures still return error.
	return resp, nil
}

func waitHealth(baseURL string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/health", nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("health status %d", res.StatusCode)
	}
	return nil
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

// mockExtension is an in-process client of the control protocol.
type mockExtension struct {
	baseURL     string
	script      string
	stopMode    string
	windowID    int
	har         json.RawMessage
	stopReason  string
	onRecording func()

	gotStart           bool
	gotStop            bool
	didHello           bool
	didRecordingStatus bool
	didComplete        bool
}

func (m *mockExtension) run(ctx context.Context) {
	switch m.script {
	case ExtNone:
		return
	case ExtHelloOnly:
		m.postHello()
		// Stay idle until ctx done.
		<-ctx.Done()
	case ExtHelloNoRecording:
		m.postHello()
		// Heartbeat with non-recording state until cancelled.
		t := time.NewTicker(50 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				m.postStatus("waiting_extension", 0)
			}
		}
	case ExtRecordAndComplete, ExtRecordNoComplete:
		m.postHello()
		// Long-poll until start command.
		if !m.waitCommand(ctx, "start") {
			return
		}
		m.gotStart = true
		m.postStatus("recording", 0)
		m.didRecordingStatus = true
		if m.onRecording != nil {
			m.onRecording()
		}

		if m.script == ExtRecordNoComplete {
			// Optionally observe stop, but never complete.
			_ = m.waitCommand(ctx, "stop")
			m.gotStop = true
			// Hang until parent cancels after complete-timeout.
			<-ctx.Done()
			return
		}

		// record-and-complete
		if m.stopMode == StopCLI {
			if !m.waitCommand(ctx, "stop") {
				return
			}
			m.gotStop = true
			if m.stopReason == "" {
				m.stopReason = "cli"
			}
		} else {
			// Extension-initiated stop: no need to wait for stop command.
			if m.stopReason == "" {
				m.stopReason = "extension"
			}
			m.gotStop = true
		}
		m.postComplete()
	default:
		// Unknown script: do nothing.
	}
}

func (m *mockExtension) postHello() {
	body := `{"version":"test-mock-1.0.0"}`
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
			// Server may be shutting down.
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
			if want == "start" {
				m.gotStart = true
			}
			if want == "stop" {
				m.gotStop = true
			}
			return true
		}
	}
}

func (m *mockExtension) postComplete() {
	har := m.har
	if len(har) == 0 {
		har = defaultMultiTabHAR()
	}
	// Accept either full HAR document or wrap entries.
	var harDoc any
	if err := json.Unmarshal(har, &harDoc); err != nil {
		harDoc = json.RawMessage(har)
	}
	reason := m.stopReason
	if reason == "" {
		reason = "extension"
	}
	payload := map[string]any{
		"har":         harDoc,
		"stop_reason": reason,
		"window_id":   m.windowID,
		"stats": map[string]any{
			"entry_count": 2,
			"tabs":        2,
		},
	}
	b, _ := json.Marshal(payload)
	res, err := http.Post(m.baseURL+"/v1/complete", "application/json", bytes.NewReader(b))
	if err != nil {
		return
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		m.didComplete = true
	}
}

func defaultMultiTabHAR() json.RawMessage {
	// Two entries from different tabs; startedDateTime out of order to force sort.
	const doc = `{
  "log": {
    "version": "1.2",
    "creator": {"name": "browser-trace-mock", "version": "1.0"},
    "entries": [
      {
        "startedDateTime": "2026-07-11T12:00:02.000Z",
        "time": 12,
        "request": {"method": "GET", "url": "https://example.com/b", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "queryString": [], "headersSize": -1, "bodySize": -1},
        "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "content": {"size": 0, "mimeType": "text/plain"}, "redirectURL": "", "headersSize": -1, "bodySize": 0},
        "cache": {},
        "timings": {"send": 0, "wait": 10, "receive": 2},
        "_tabId": 2
      },
      {
        "startedDateTime": "2026-07-11T12:00:01.000Z",
        "time": 10,
        "request": {"method": "GET", "url": "https://example.com/a", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "queryString": [], "headersSize": -1, "bodySize": -1},
        "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "content": {"size": 0, "mimeType": "text/plain"}, "redirectURL": "", "headersSize": -1, "bodySize": 0},
        "cache": {},
        "timings": {"send": 0, "wait": 8, "receive": 2},
        "_tabId": 1
      }
    ]
  }
}`
	return json.RawMessage(doc)
}
```
