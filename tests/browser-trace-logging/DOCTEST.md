# browser-trace lifecycle logging

Exercises **progress / lifecycle logging** for `browser-trace` / package
`browsertrace`: default **stderr info milestones**, **quiet**, **verbose**,
ready-wait **heartbeat**, rich **ready-timeout** messages, and optional session
**log file** — while keeping success **stdout** contract stable for scripts.

Tests drive **`browsertrace.Run`** in-process with buffer `Stdout`/`Stderr`, a
**mock extension** over HTTP, short timeouts, and injectable
`ReadyHeartbeat` (product default ~5s). **No real Chrome**.

Product defaults under test (logging-related):

| Setting | Default | Notes |
|---------|---------|-------|
| stderr level | **info** milestones | not quiet |
| stdout on success | **only** session dir path + `\n` | scripts/tests contract |
| ready heartbeat | every ~5s | injectable via `Config.ReadyHeartbeat` |
| ready timeout message | stage + session URL + install hint | errors on stderr |
| `-v` / `Verbose` | off | hello, start recording, stop, complete |
| `Quiet` | off | suppress info; errors still on stderr |
| session log file | `{sessionDir}/browser-trace.log` | off when Quiet or `NoLogFile` |

## Version

0.0.2

# DSN (Domain Specific Notion)

**User** runs **browser-trace** and watches **stderr** for progress while
**stdout** stays machine-readable (session path only on success).

**Lifecycle Logger** is the logging facet of **browser-trace**:

- Emits **info milestones** on stderr by default: listen/addr, session URL or
  id, ready-wait (with timeout), ready **heartbeat** while waiting, recording
  started, optional saved note.
- Honors **Quiet** (no info milestones; errors still stderr) and **Verbose**
  (extra detail: hello/version, start recording, stop, complete).
- Mirrors info+ lines into **`{sessionDir}/browser-trace.log`** unless Quiet or
  `NoLogFile`.
- On **ready timeout**, prints a **rich failure** on stderr: timeout language,
  stage (`no_hello` / `no_recording`), session URL or `/go?session=`, install
  path hint when known.

**Control Server** and **Mock Extension** behave as in the core session tree:
hello → status recording → complete (or silent for timeout paths). Logging is
orthogonal to bind/HAR save contracts; success **stdout** remains
`{sessionDir}\n` only.

**Storage** under `{BaseDir}/YYYY-MM-DD-HH-MM-SS-<suffix>/`:

- `meta.json`, `recording.har` (existing)
- `browser-trace.log` (**new** when logging to file is enabled)

## Decision Tree

```
browser-trace lifecycle logging
├── success/                              [mock record+complete; exit 0]
│   ├── default-info/                     [Quiet=false, Verbose=false]
│   │   ├── with-log-file/                [NoLogFile=false — default]
│   │   │   milestones + clean stdout + browser-trace.log
│   │   └── no-log-file/                  [NoLogFile=true]
│   │       milestones on stderr; no browser-trace.log
│   ├── quiet/                            [Quiet=true]
│   │   └── no-info-milestones            stderr free of info tokens; stdout path\n
│   └── verbose/                          [Verbose=true]
│       └── hello-and-milestones          stderr mentions hello/version (+ info ok)
└── ready-fail/                           [never recording within ReadyTimeout]
    └── no-hello/                         [mock silent]
        ├── rich-timeout-message          timeout + stage + session URL + install hint
        └── ready-heartbeat               short ReadyHeartbeat; ≥2 wait heartbeats
```

### Parameter significance (high → low)

1. **Session outcome** — success vs ready-fail changes exit code and log sequence.
2. **Log mode** (success) — default info vs quiet vs verbose (MECE verbosity).
3. **Log file policy** (default-info) — write `browser-trace.log` vs `NoLogFile`.
4. **Ready-fail detail** (no-hello) — rich final message vs heartbeat while waiting.

### MECE notes

- **Quiet** and **Verbose** are separate branches under `success/`. When both
  flags are set in product code, **Quiet wins** (no info, no verbose extras);
  this tree does not add a combined leaf.
- Heartbeat is only asserted on a deliberate long ready-wait with injectable
  short `ReadyHeartbeat` (not on the short rich-timeout leaf).

## Test Index

| Leaf | Scenario (requirement) |
|------|------------------------|
| `success/default-info/with-log-file` | (#1) Default success: stderr milestones + stdout path only + log file |
| `success/default-info/no-log-file` | (#1 + NoLogFile) Milestones on stderr; no `browser-trace.log` |
| `success/quiet/no-info-milestones` | (#2) Quiet: no info milestones; stdout still path + `\n` |
| `success/verbose/hello-and-milestones` | (#4) Verbose: hello/version on stderr |
| `ready-fail/no-hello/rich-timeout-message` | (#3) Ready timeout rich stderr |
| `ready-fail/no-hello/ready-heartbeat` | Heartbeat while waiting (injectable interval) |

Regression: existing `./tests/browser-trace` remains GREEN for stdout contracts
(stderr may grow with new info lines; those leaves assert stdout/files, not empty stderr).

## How to Run

```sh
doctest vet ./tests/browser-trace-logging
doctest test ./tests/browser-trace-logging
# regression (stdout contracts)
doctest test ./tests/browser-trace
```

### Expected `browsertrace` API (logging contract)

Package `github.com/xhd2015/browser-agent/browsertrace` extends
`Config` / `Run` with:

- `Verbose bool` — extra detail (hello, start recording, stop, complete)
- `Quiet bool` — suppress info milestones; errors still on stderr
- `NoLogFile bool` — if true, stderr only (no `{sessionDir}/browser-trace.log`)
- `ReadyHeartbeat time.Duration` — ready-wait heartbeat interval; **default 5s**;
  tests inject short values (e.g. 50ms)
- existing: `Addr`, `BaseDir`, `ReadyTimeout`, `CompleteTimeout`, `NoOpenChrome`,
  `SessionSuffix`, `Stdout`, `Stderr`

**Stdout success contract (stable):** on exit 0, stdout is **exactly** the
session directory absolute path followed by a single trailing newline — no
extra banner lines.

**Stderr info milestones (default, not Quiet)** — match **tokens/substrings**,
not a rigid full string (prefix may be `browser-trace:` or timestamped):

1. listen / listening / addr
2. session URL or session id
3. ready waiting (timeout duration mentioned)
4. ready heartbeat when wait ≥ heartbeat interval and still not ready
   (tokens such as waiting / left / `no_hello` / `no_recording`)
5. recording started (when status recording)
6. optional “saved” on success path

**Ready timeout stderr** must include:

- timeout / ready failure language
- stage hint (`no_hello` or connect/hello language; `no_recording` when applicable)
- session URL or `/go?session=`
- install path hint when extract path is known

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
	ExtNone              = "none"
	ExtHelloOnly         = "hello-only"
	ExtHelloNoRecording  = "hello-no-recording"
	ExtRecordAndComplete = "record-and-complete"
)

// StopMode values.
const (
	StopNone      = "none"
	StopExtension = "extension"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Addr            string
	BaseDir         string
	ReadyTimeout    time.Duration
	CompleteTimeout time.Duration
	NoOpenChrome    bool
	SessionSuffix   string

	// Logging / verbosity (product Config fields under test).
	Verbose   bool
	Quiet     bool
	NoLogFile bool
	// ReadyHeartbeat is injectable ready-wait heartbeat interval.
	// Zero → product default (5s). Tests use short values (e.g. 50ms).
	ReadyHeartbeat time.Duration

	// ExtensionScript selects mock extension behavior after the server is up.
	ExtensionScript string
	// StopMode selects stop initiator once recording is observed.
	StopMode string

	MockHAR        json.RawMessage
	MockStopReason string
	MockWindowID   int
}

// Response is collected after browsertrace.Run returns.
type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
	ErrText  string

	SessionDir string
	MetaPath   string
	HARPath    string
	MetaJSON   []byte
	HARJSON    []byte

	// Log file side effect
	LogFilePath   string
	LogFileExists bool
	LogFile       []byte

	// Mock observations
	HelloOK         bool
	StatusRecording bool
	CompletePosted  bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	req.NoOpenChrome = true
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

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()
	mockCtx, mockCancel := context.WithCancel(context.Background())
	defer mockCancel()

	var mockWG sync.WaitGroup
	if mock != nil {
		mockWG.Add(1)
		go func() {
			defer mockWG.Done()
			deadline := time.Now().Add(req.ReadyTimeout + 2*time.Second)
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
		Addr:            addr,
		BaseDir:         req.BaseDir,
		ReadyTimeout:    req.ReadyTimeout,
		CompleteTimeout: req.CompleteTimeout,
		NoOpenChrome:    true,
		SessionSuffix:   req.SessionSuffix,
		Stdout:          &stdout,
		Stderr:          &stderr,
		Verbose:         req.Verbose,
		Quiet:           req.Quiet,
		NoLogFile:       req.NoLogFile,
		ReadyHeartbeat:  req.ReadyHeartbeat,
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
		resp.LogFilePath = filepath.Join(resp.SessionDir, "browser-trace.log")
		if b, err := os.ReadFile(resp.MetaPath); err == nil {
			resp.MetaJSON = b
		}
		if b, err := os.ReadFile(resp.HARPath); err == nil {
			resp.HARJSON = b
		}
		if b, err := os.ReadFile(resp.LogFilePath); err == nil {
			resp.LogFileExists = true
			resp.LogFile = b
		}
	}
	if mock != nil {
		resp.HelloOK = mock.didHello
		resp.CompletePosted = mock.didComplete
		if mock.didRecordingStatus {
			resp.StatusRecording = true
		}
	}

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
		<-ctx.Done()
	case ExtHelloNoRecording:
		m.postHello()
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
	case ExtRecordAndComplete:
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
		if m.stopReason == "" {
			m.stopReason = "extension"
		}
		m.gotStop = true
		m.postComplete()
	default:
	}
}

func (m *mockExtension) postHello() {
	body := `{"version":"test-mock-1.0.0","features":["browser-trace"]}`
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

func (m *mockExtension) postComplete() {
	har := m.har
	if len(har) == 0 {
		har = defaultMiniHAR()
	}
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
			"entry_count": 1,
			"tabs":        1,
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

func defaultMiniHAR() json.RawMessage {
	const doc = `{
  "log": {
    "version": "1.2",
    "creator": {"name": "browser-trace-mock", "version": "1.0"},
    "entries": [
      {
        "startedDateTime": "2026-07-11T12:00:01.000Z",
        "time": 10,
        "request": {"method": "GET", "url": "https://example.com/a", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "queryString": [], "headersSize": -1, "bodySize": -1},
        "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "content": {"size": 0, "mimeType": "text/plain"}, "redirectURL": "", "headersSize": -1, "bodySize": 0},
        "cache": {},
        "timings": {"send": 0, "wait": 8, "receive": 2}
      }
    ]
  }
}`
	return json.RawMessage(doc)
}
```
