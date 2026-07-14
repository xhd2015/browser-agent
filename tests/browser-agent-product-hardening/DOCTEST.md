# browser-agent product hardening — manifest permission contract + session info control vs browser

Seals two product gaps in package
`github.com/xhd2015/browser-agent/browseragent`:

| Surface | What is under test |
|---------|-------------------|
| Manifest validate | Pure JSON + production FS manifests require `debugger`, `tabs`, `alarms`, `storage`, port **43761** hosts, and broad host access |
| Session info CLI | `session info` always prints control snapshot; merges browser `info` job (tabs) only when extension connected; never fabricates tabs |

**No production code in this tree.** **No real Chrome.** Fake WebSocket
extension only for the connected session-info leaf.

Related regressions (run after implement, not part of this tree):

```sh
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-session-nested/...
doctest test ./tests/browser-agent-cli-react/...
```

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** (or **Agent**) uses **browser-agent** package surfaces:

### 1. Extension Manifest Contract

**Production extension** is an MV3 Chrome package. Two on-disk sources must
satisfy the same permission contract:

1. **Embedded package** — `browseragent/embedded/extension/manifest.json`
   (what `ExtractEmbeddedExtension` / install ships).
2. **Product shell** — `Chrome-Ext-Browser-Agent/public/manifest.json`
   (module-root source of truth for the extension product).

**Required `permissions` (all present):**

- `debugger` — CDP attach
- `tabs` — tab inventory / targeting
- `alarms` — keep-alive / reconnect
- `storage` — local state

**Required host coverage:**

- Control plane: `http://127.0.0.1:43761/*` and/or `http://localhost:43761/*`
- Broad page access for CDP: `<all_urls>` **or** an equivalent broad host
  pattern needed to attach on arbitrary pages

**Validator** (pure helper preferred):

```text
ValidateExtensionManifestJSON([]byte) error
  // optional twin: ValidateExtensionManifestPath(path) error
```

- Valid → `nil`
- Missing required permission → non-nil error whose text **mentions** the
  missing name (e.g. `debugger`, `tabs`)
- Missing required host coverage → non-nil error mentioning host / 43761 /
  `<all_urls>` as appropriate

### 2. Session info CLI (control vs browser)

```text
browser-agent session info [--session-id] [--addr]
```

**Session id resolve** (same as other side-commands):

1. `--session-id` when set  
2. else env `BROWSER_AGENT_SESSION_ID`  
3. else error mentioning **both** `--session-id` and `BROWSER_AGENT_SESSION_ID`

**Behavior when session resolves:**

1. Always `GET /v1/session` (control snapshot: phase, `extension.connected`,
   install path, hint, …).
2. **If `extension.connected == true`:** also enqueue job type `info`
   (`POST /v1/jobs`, wait) and merge useful browser fields into stdout
   (especially `tabs` when present).
3. **If not connected:** do **not** invent tabs; stdout still shows control
   status clearly and indicates browser tab inventory needs a connected
   extension (`browser` null / `browser_error` / equivalent).
4. Prefer pretty JSON stdout ending with `\n`. Extra top-level keys OK.
   Must distinguish **control** vs **browser** data.
5. Control fetch failure → non-zero / non-nil error. Browser job failure while
   connected → prefer exit 0 with `browser_error` (partial success) so
   operators still see control status (optional hardening; not a separate
   leaf here).

**Illustrative shapes** (field names flexible if tests document them):

```json
{
  "session_id": "...",
  "phase": "waiting_extension",
  "extension": { "connected": false },
  "browser": null,
  "browser_error": "extension not connected"
}
```

```json
{
  "session_id": "...",
  "phase": "extension_connected",
  "extension": { "connected": true },
  "browser": { "tabs": [ { "id": 1, "url": "https://example.com/" } ] }
}
```

**Test Client** in this tree:

- Pure leaves call `ValidateExtensionManifestJSON` on fixture bytes.
- FS leaves read ModuleRoot manifests and validate the same helper.
- CLI leaves call `HandleCLI` with injectable env + buffers; live leaves start
  `browseragent.Run` (NoOpenChrome, NoAgentRun) and optionally a **fake
  extension** WebSocket that auto-completes `info` jobs with tabs.

## Decision Tree

```
browser-agent-product-hardening
├── manifest-validate/                         [ValidateExtensionManifestJSON]
│   ├── pure-json/                               fixture bytes
│   │   ├── valid-all-required/                    A1 all perms + hosts → ok
│   │   ├── missing-debugger/                      A2 error mentions debugger
│   │   └── missing-tabs/                          A3 error mentions tabs
│   ├── embedded-extension/
│   │   └── satisfies-required/                    A4 embedded manifest ok
│   └── shell-public/
│       └── satisfies-required/                    A5 public/manifest.json ok
└── session-info-cli/                          [HandleCLI session info]
    ├── control-only-disconnected/                 B1 no WS; control only; no fake tabs
    ├── connected-with-info-tabs/                  B2 fake WS + info job tabs merged
    └── missing-session-error/                     B3 no sid → flag+env error
```

### Parameter significance (high → low)

1. **Surface / Mode** — manifest validate vs session-info CLI (different
   contracts; one Mode per top branch).
2. **Within manifest** — pure fixture vs production FS source (bytes vs path).
3. **Within pure** — valid vs which required permission is missing.
4. **Within session-info** — resolve error vs live disconnected vs live
   connected+info (largest outcome split).

## Test Index

| Leaf | Scenario |
|------|----------|
| `manifest-validate/pure-json/valid-all-required` | (A1) fixture with debugger+tabs+alarms+storage + 43761 hosts + broad host → Validate OK |
| `manifest-validate/pure-json/missing-debugger` | (A2) fixture missing `debugger` → error mentions debugger |
| `manifest-validate/pure-json/missing-tabs` | (A3) fixture missing `tabs` → error mentions tabs |
| `manifest-validate/embedded-extension/satisfies-required` | (A4) `browseragent/embedded/extension/manifest.json` validates OK |
| `manifest-validate/shell-public/satisfies-required` | (A5) `Chrome-Ext-Browser-Agent/public/manifest.json` validates OK |
| `session-info-cli/control-only-disconnected` | (B1) live serve, no WS → JSON control waiting/connected false; no fabricated page tabs; browser unavailable signal; trailing `\n` |
| `session-info-cli/connected-with-info-tabs` | (B2) fake WS hello + info job with tabs → connected true; stdout includes browser/tabs |
| `session-info-cli/missing-session-error` | (B3) `session info` without sid/env → error mentions `--session-id` and `BROWSER_AGENT_SESSION_ID` |

**Leaf count: 8**

## How to Run

```sh
cd project-api-capture
doctest vet ./tests/browser-agent-product-hardening
doctest test ./tests/browser-agent-product-hardening/...
# regressions after implement:
# doctest test ./tests/browser-agent/...
# doctest test ./tests/browser-agent-session-nested/...
# doctest test ./tests/browser-agent-cli-react/...
```

### Implementer contract (authoritative for GREEN)

```text
// Manifest (pure preferred)
func ValidateExtensionManifestJSON(data []byte) error
// optional:
// func ValidateExtensionManifestPath(path string) error

// Required permissions: debugger, tabs, alarms, storage
// Required hosts: http://127.0.0.1:43761/* and/or http://localhost:43761/*
//                 plus <all_urls> or equivalent broad host access

// Session info CLI (via existing HandleCLI)
// browser-agent session info [--session-id] [--addr]
// 1. GET /v1/session always
// 2. if extension.connected: POST /v1/jobs type=info, merge browser/tabs
// 3. if not connected: no fabricated tabs; signal browser unavailable
// 4. pretty JSON stdout + trailing \n
// 5. missing session id → error mentions --session-id and BROWSER_AGENT_SESSION_ID

func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode values — top-level surface under test.
const (
	ModeManifestValidate = "manifest-validate"
	ModeSessionInfoCLI   = "session-info-cli"
)

// ManifestSource for ModeManifestValidate.
const (
	ManifestSourceBytes    = "bytes"    // pure fixture / inline JSON
	ManifestSourceEmbedded = "embedded" // browseragent/embedded/extension/manifest.json
	ManifestSourceShell    = "shell"    // Chrome-Ext-Browser-Agent/public/manifest.json
)

// SessionInfoKind for ModeSessionInfoCLI.
const (
	SessionInfoControlOnlyDisconnected = "control-only-disconnected"
	SessionInfoConnectedWithInfoTabs   = "connected-with-info-tabs"
	SessionInfoMissingSession          = "missing-session-error"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// ModuleRoot is project-api-capture module directory (filesystem leaves).
	// Root Setup sets from DOCTEST_ROOT/../..
	ModuleRoot string

	// --- manifest-validate ---
	ManifestSource string // bytes | embedded | shell
	ManifestJSON   []byte // when Source=bytes (or override)
	ManifestPath   string // optional explicit path; FS sources fill this

	// --- session-info-cli ---
	SessionInfoKind string
	CLIArgs         []string
	CLIEnv          map[string]string
	MaxDispatchWait time.Duration

	// Live serve
	BaseDir       string
	Addr          string
	SessionID     string
	NoOpenChrome  bool
	NoAgentRun    bool
	ReadyTimeout  time.Duration
	FakeExtension bool

	// Fake info job payload (connected leaf)
	InfoJobTabs []map[string]any
	InfoVersion string
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// Manifest
	ValidateErr     string
	ValidateOK      bool
	ManifestPath    string
	ManifestText    string
	ManifestSource  string

	// CLI
	Stdout           string
	Stderr           string
	ExitCode         int
	ErrText          string
	CLIErr           string
	DispatchTimedOut bool

	// Live / parsed helpers
	BaseURL           string
	RealSessionID     string
	ParsedJSON        map[string]any
	ExtensionConnected bool
	Phase             string
	HasTabsKey        bool
	TabCount          int
	BrowserUnavailable bool
	WSJobReceived     bool
	ObservedJobType   string
	JobsSeen          []map[string]any
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}
	switch req.Mode {
	case ModeManifestValidate:
		return runManifestValidate(t, req)
	case ModeSessionInfoCLI:
		return runSessionInfoCLI(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- manifest validate ---

func runManifestValidate(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	src := req.ManifestSource
	if src == "" {
		if len(req.ManifestJSON) > 0 {
			src = ManifestSourceBytes
		} else if req.ManifestPath != "" {
			src = ManifestSourceBytes // path override treated as path read
		} else {
			t.Fatal("ManifestSource or ManifestJSON/ManifestPath must be set")
		}
	}

	data := req.ManifestJSON
	path := req.ManifestPath
	switch src {
	case ManifestSourceBytes:
		if len(data) == 0 && path != "" {
			b, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read manifest path %s: %w", path, err)
			}
			data = b
		}
		if len(data) == 0 {
			t.Fatal("ManifestJSON empty for pure bytes source")
		}
	case ManifestSourceEmbedded:
		if path == "" {
			path = filepath.Join(req.ModuleRoot, "browseragent", "embedded", "extension", "manifest.json")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read embedded manifest %s: %w", path, err)
		}
		data = b
	case ManifestSourceShell:
		if path == "" {
			// Prefer public/, then common fallbacks.
			cands := []string{
				filepath.Join(req.ModuleRoot, "Chrome-Ext-Browser-Agent", "public", "manifest.json"),
				filepath.Join(req.ModuleRoot, "Chrome-Ext-Browser-Agent", "manifest.json"),
				filepath.Join(req.ModuleRoot, "Chrome-Ext-Browser-Agent", "build", "manifest.json"),
			}
			var err error
			for _, c := range cands {
				var b []byte
				b, err = os.ReadFile(c)
				if err == nil {
					path = c
					data = b
					break
				}
			}
			if len(data) == 0 {
				return nil, fmt.Errorf("shell manifest not found under Chrome-Ext-Browser-Agent: %v", err)
			}
		} else {
			b, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read shell manifest %s: %w", path, err)
			}
			data = b
		}
	default:
		return nil, fmt.Errorf("unknown ManifestSource %q", src)
	}

	resp := &Response{
		ManifestPath:   path,
		ManifestText:   string(data),
		ManifestSource: src,
	}

	// Production helper under test (RED until implementer lands it).
	err := browseragent.ValidateExtensionManifestJSON(data)
	if err != nil {
		resp.ValidateOK = false
		resp.ValidateErr = err.Error()
		resp.ErrText = err.Error()
		resp.ExitCode = 1
		return resp, nil
	}
	resp.ValidateOK = true
	resp.ExitCode = 0
	return resp, nil
}

// --- session info CLI ---

func runSessionInfoCLI(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	kind := req.SessionInfoKind
	if kind == "" && len(req.CLIArgs) == 0 {
		t.Fatal("SessionInfoKind or CLIArgs must be set")
	}
	if kind == SessionInfoMissingSession {
		if len(req.CLIArgs) == 0 {
			req.CLIArgs = []string{"session", "info"}
		}
		if req.CLIEnv == nil {
			req.CLIEnv = map[string]string{}
		}
		return invokeHandleCLI(t, req)
	}
	if kind == SessionInfoControlOnlyDisconnected {
		return runSessionInfoLive(t, req, false)
	}
	if kind == SessionInfoConnectedWithInfoTabs {
		return runSessionInfoLive(t, req, true)
	}
	// Fallback: raw CLI args only
	return invokeHandleCLI(t, req)
}

func runSessionInfoLive(t *testing.T, req *Request, withFakeExt bool) (*Response, error) {
	t.Helper()
	req.NoOpenChrome = true
	req.NoAgentRun = true
	if req.MaxDispatchWait <= 0 {
		req.MaxDispatchWait = 10 * time.Second
	}
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}

	srv, cleanup, err := startAgentServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var mu sync.Mutex
	resp := &Response{
		BaseURL:       srv.BaseURL,
		RealSessionID: srv.SessionID,
	}

	if withFakeExt {
		tabs := req.InfoJobTabs
		if len(tabs) == 0 {
			tabs = []map[string]any{
				{"id": 1, "url": "https://example.com/", "title": "Example"},
				{"id": 2, "url": "https://shop.example/", "title": "Shop"},
			}
		}
		ver := req.InfoVersion
		if ver == "" {
			ver = "1.0.0"
		}
		ext, err := dialFakeExtension(srv.BaseURL, ver, []string{"browser-agent"})
		if err != nil {
			return resp, fmt.Errorf("fake extension dial: %w", err)
		}
		defer ext.Close()
		ext.AutoCompleteOK = true
		ext.ResultData = map[string]any{
			"tabs":     tabs,
			"version":  ver,
			"features": []any{"browser-agent"},
		}
		// Also accept nested data shapes the implementer may forward as-is.
		ext.OnJob = func(jobType string, params map[string]any) {
			mu.Lock()
			defer mu.Unlock()
			resp.WSJobReceived = true
			resp.ObservedJobType = jobType
			resp.JobsSeen = append(resp.JobsSeen, map[string]any{
				"type":   jobType,
				"params": params,
			})
		}
		go ext.Loop()
		// Brief settle so hello is processed before CLI.
		time.Sleep(50 * time.Millisecond)
	}

	args := req.CLIArgs
	if len(args) == 0 {
		args = []string{
			"session", "info",
			"--session-id", srv.SessionID,
			"--addr", srv.BaseURL,
			"--json",
		}
	} else {
		args = injectAddrAndSession(args, srv.BaseURL, srv.SessionID)
	}

	req2 := *req
	req2.CLIArgs = args
	cliResp, err := invokeHandleCLI(t, &req2)
	if cliResp != nil {
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
		resp.CLIErr = cliResp.CLIErr
		resp.ErrText = cliResp.ErrText
		resp.DispatchTimedOut = cliResp.DispatchTimedOut
	}
	fillParsedSessionInfo(resp)

	if withFakeExt {
		// Wait briefly for info job observation (CLI may already have finished).
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			mu.Lock()
			got := resp.WSJobReceived
			mu.Unlock()
			if got {
				break
			}
			time.Sleep(15 * time.Millisecond)
		}
	}
	return resp, err
}

func fillParsedSessionInfo(resp *Response) {
	if resp == nil || strings.TrimSpace(resp.Stdout) == "" {
		return
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Stdout)), &m); err != nil {
		return
	}
	resp.ParsedJSON = m
	if phase, ok := m["phase"].(string); ok {
		resp.Phase = phase
	}
	if ext, ok := m["extension"].(map[string]any); ok {
		if c, ok := ext["connected"].(bool); ok {
			resp.ExtensionConnected = c
		}
	}
	// connected may also appear at top level in older snapshots
	if c, ok := m["connected"].(bool); ok && !resp.ExtensionConnected {
		resp.ExtensionConnected = c
	}

	// Detect tabs anywhere useful under browser / data / top-level.
	tabs := findTabsSlice(m)
	if tabs != nil {
		resp.HasTabsKey = true
		resp.TabCount = len(tabs)
	}

	// Browser unavailable signals
	if be, ok := m["browser_error"].(string); ok && strings.TrimSpace(be) != "" {
		resp.BrowserUnavailable = true
	}
	if b, ok := m["browser"]; ok && b == nil {
		resp.BrowserUnavailable = true
	}
	low := strings.ToLower(resp.Stdout)
	if strings.Contains(low, "not connected") ||
		strings.Contains(low, "extension not connected") ||
		strings.Contains(low, "browser unavailable") ||
		strings.Contains(low, "requires a connected extension") {
		resp.BrowserUnavailable = true
	}
}

func findTabsSlice(m map[string]any) []any {
	if m == nil {
		return nil
	}
	// browser.tabs
	if b, ok := m["browser"].(map[string]any); ok {
		if t, ok := b["tabs"].([]any); ok {
			return t
		}
		if d, ok := b["data"].(map[string]any); ok {
			if t, ok := d["tabs"].([]any); ok {
				return t
			}
		}
	}
	// top-level tabs (discouraged but accepted if documented)
	if t, ok := m["tabs"].([]any); ok {
		return t
	}
	if d, ok := m["data"].(map[string]any); ok {
		if t, ok := d["tabs"].([]any); ok {
			return t
		}
	}
	return nil
}

func invokeHandleCLI(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	maxWait := req.MaxDispatchWait
	if maxWait <= 0 {
		maxWait = 3 * time.Second
	}
	var stdout, stderr bytes.Buffer
	env := req.CLIEnv
	if env == nil {
		env = map[string]string{}
	}
	args := req.CLIArgs
	if args == nil {
		args = []string{}
	}

	type outcome struct {
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		done <- outcome{err: browseragent.HandleCLI(args, env, &stdout, &stderr)}
	}()

	resp := &Response{}
	select {
	case out := <-done:
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		if out.err != nil {
			resp.CLIErr = out.err.Error()
			resp.ErrText = out.err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		resp.ExitCode = 0
		return resp, nil
	case <-time.After(maxWait):
		resp.DispatchTimedOut = true
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.ExitCode = 1
		resp.ErrText = "HandleCLI timed out (possible accidental serve hang)"
		return resp, fmt.Errorf("%s", resp.ErrText)
	}
}

func injectAddrAndSession(args []string, baseURL, sessionID string) []string {
	hasAddr, hasSession := false, false
	for _, a := range args {
		if a == "--addr" || strings.HasPrefix(a, "--addr=") {
			hasAddr = true
		}
		if a == "--session-id" || strings.HasPrefix(a, "--session-id=") {
			hasSession = true
		}
	}
	out := append([]string{}, args...)
	insertAt := 0
	if len(out) >= 2 && out[0] == "session" {
		insertAt = 2
	} else if len(out) >= 1 {
		insertAt = 1
	}
	if !hasSession {
		ins := []string{"--session-id", sessionID}
		out = append(out[:insertAt:insertAt], append(ins, out[insertAt:]...)...)
		insertAt += 2
	}
	if !hasAddr {
		ins := []string{"--addr", baseURL}
		out = append(out[:insertAt:insertAt], append(ins, out[insertAt:]...)...)
	}
	return out
}

// --- serve + fake extension ---

type agentServer struct {
	BaseURL   string
	SessionID string
	BaseDir   string
	cancel    context.CancelFunc
}

func startAgentServer(t *testing.T, req *Request) (*agentServer, func(), error) {
	t.Helper()
	baseDir := req.BaseDir
	if baseDir == "" {
		var err error
		baseDir, err = os.MkdirTemp("", "ba-product-hard-*")
		if err != nil {
			return nil, nil, err
		}
	}
	sid := req.SessionID
	if sid == "" {
		sid = fmt.Sprintf("sess-ph-%d", time.Now().UnixNano()%1e12)
	}
	addr := req.Addr
	if addr == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, nil, err
		}
		addr = ln.Addr().String()
		_ = ln.Close()
	}
	readyTO := req.ReadyTimeout
	if readyTO <= 0 {
		readyTO = 5 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      baseDir,
		SessionID:    sid,
		NoOpenChrome: true,
		NoAgentRun:   true,
		ReadyTimeout: readyTO,
		Stdout:       io.Discard,
		Stderr:       io.Discard,
	}
	errCh := make(chan error, 1)
	go func() {
		_, err := browseragent.Run(ctx, cfg)
		errCh <- err
	}()
	baseURL := "http://" + addr
	if err := waitHealth(baseURL, readyTO); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("serve health: %w", err)
	}
	srv := &agentServer{BaseURL: baseURL, SessionID: sid, BaseDir: baseDir, cancel: cancel}
	cleanup := func() {
		cancel()
		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
		}
		_ = os.RemoveAll(baseDir)
	}
	return srv, cleanup, nil
}

func waitHealth(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		resp, err := http.Get(strings.TrimRight(baseURL, "/") + "/v1/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
			last = fmt.Errorf("health status %d", resp.StatusCode)
		} else {
			last = err
		}
		time.Sleep(25 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("health timeout")
	}
	return last
}

type fakeExtension struct {
	conn           *websocket.Conn
	AutoCompleteOK bool
	ResultData     map[string]any
	OnJob          func(jobType string, params map[string]any)
	JobsSeen       []map[string]any
	mu             sync.Mutex
}

func dialFakeExtension(baseURL, version string, features []string) (*fakeExtension, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	ext := &fakeExtension{conn: conn}
	hello := map[string]any{
		"v":    1,
		"type": "hello",
		"payload": map[string]any{
			"version":  version,
			"features": features,
		},
	}
	if err := conn.WriteJSON(hello); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return ext, nil
}

func (f *fakeExtension) Close() {
	if f.conn != nil {
		_ = f.conn.Close()
	}
}

func (f *fakeExtension) Loop() {
	for {
		var msg map[string]any
		if err := f.conn.ReadJSON(&msg); err != nil {
			return
		}
		typ, _ := msg["type"].(string)
		if typ != "job" {
			continue
		}
		payload, _ := msg["payload"].(map[string]any)
		if payload == nil {
			payload, _ = msg["job"].(map[string]any)
		}
		jobID := stringField(payload, "id", "job_id", "jobId")
		if jobID == "" {
			jobID = stringField(msg, "id", "job_id", "jobId")
		}
		jobType := stringField(payload, "type", "job_type", "jobType")
		params, _ := payload["params"].(map[string]any)
		f.mu.Lock()
		f.JobsSeen = append(f.JobsSeen, map[string]any{"type": jobType, "params": params})
		onJob := f.OnJob
		f.mu.Unlock()
		if onJob != nil {
			onJob(jobType, params)
		}
		if f.AutoCompleteOK {
			result := map[string]any{
				"v":    1,
				"type": "result",
				"payload": map[string]any{
					"job_id": jobID,
					"ok":     true,
					"data":   f.ResultData,
				},
			}
			_ = f.conn.WriteJSON(result)
		}
	}
}

func stringField(m map[string]any, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			switch t := v.(type) {
			case string:
				if t != "" {
					return t
				}
			default:
				s := fmt.Sprint(t)
				if s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}

// silence unused imports in some codegen paths
var (
	_ = bytes.Buffer{}
	_ = json.Marshal
	_ = filepath.Join
	_ = sync.Mutex{}
	_ = net.Listen
	_ = url.Parse
)
```
