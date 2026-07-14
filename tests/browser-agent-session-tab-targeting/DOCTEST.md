# browser-agent session tab targeting (--tab-id / --tab-index)

Classic TDD for first-class **tab targeting** on session job commands (`eval`, `run`,
`logs`, `screenshot`, `cdp`) and richer **`session info`** output (human table +
`--json` tab index / `job_target`).

| Surface | What is under test |
|---------|-------------------|
| CLI flags | `--tab-id` posts `tab_id` in job JSON; `--tab-id` + `--tab-index` mutual exclusion |
| Extension source | Explicit `tab_id` window validation; 1-based `tab_index` order; attach reuse + detach on switch |
| session info | Human columns Idx/ID/Role; `--json` `tabs[].index` + `job_target.tab_index` |
| E2E (playwright-debug) | `--tab-id` hits background tab without focus; eval + screenshot same tab |

**No real Chrome** for cli / ext-source / info leaves. E2e uses `playwright-debug`
+ embedded extension.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Session Window** — Chrome window where `/go?session=<id>` is open. Extension
`sessions` map entry stores `{ ws, tabId, windowId, … }`.

**Tab Targeting Policy** (highest priority first):

1. `--tab-id <chromeTabId>` when set (strongly recommended for agents)
2. `--tab-index <n>` else (1-based capturable tabs in session window, left→right)
3. Active capturable tab in `entry.windowId`
4. Session page tab fallback

**Operator CLI** resolves `--tab-index` → `tab_id` before `POST /v1/jobs` (preferred).
Mutual exclusion: both `--tab-id` and `--tab-index` → exit 1 before job POST.
`--tab-index` emits stderr warning preferring `--tab-id`.

**Job payload** (`POST /v1/jobs`):

```json
{ "session_id": "sess-xxx", "type": "eval", "tab_id": 216771574, "params": { "expression": "..." } }
```

**Background Worker** validates `tab_id` belongs to session window, resolves
1-based index over capturable tabs, reuses `chrome.debugger` attach on same tab,
detaches when switching `tab_id`, serializes attach per session.

**session info** — human table (Idx | ID | Active | Role | Title) + footer job
target line; `--json` adds `tabs[].index`, `job_target.tab_index`, `recommended_cli`.

**Daemon Host** (`RunDaemon`) binds loopback, serves `/v1/sessions`, `/v1/jobs`,
`/v1/session`, `/go`. **Test Client** creates sessions via `POST /v1/sessions`
(no `openChrome`). **Fake Extension** dials `/v1/ws?session=<id>` for CLI/info
leaves. **Playwright Harness** (e2e) extracts embedded extension and runs leaf
`testdata/*.js`.

```text
HandleCLI session eval --tab-id 222 --session-id S
  -> POST /v1/jobs { tab_id: 222, type: eval, … }
  -> WS job envelope includes tab_id
  -> extension attaches debugger to tab 222 (not active tab)

session info --json
  -> tabs[{ index, id, role, active, … }]
  -> job_target { tab_id, tab_index, reason }
```

## Decision Tree

```
browser-agent-session-tab-targeting
├── cli/                              [CLI flag parsing + mutual exclusion]
│   ├── tab-id-flag-posts-payload/      --tab-id → tab_id in WS job envelope
│   └── tab-id-index-conflict/            both flags → exit 1 before POST
├── ext-source/                       [extension background.js contract]
│   ├── resolve-tab-id-window/            tab_id + entry.windowId validation
│   ├── resolve-tab-index-order/          1-based capturable tab index in window
│   └── attach-reuse-same-tab/            reuse attach; detach on tab switch
├── info/                             [session info human + json]
│   ├── json-tab-index-field/             --json tabs[].index + job_target.tab_index
│   └── human-table-columns/              human full table + job_target/recommended/hint footer
└── e2e/                              [playwright, slow+ui-automation]
    ├── eval-tab-id-background/           eval --tab-id hits unfocused tab
    └── eval-then-screenshot-same-tab/    eval + screenshot same --tab-id
```

### Parameter significance (high → low)

1. **Surface** — CLI vs ext-source vs info vs e2e.
2. **Targeting input** — explicit `--tab-id` vs `--tab-index` vs default/active.
3. **Within ext-source** — window validation vs index order vs attach lifecycle.
4. **Within info** — `--json` machine fields vs human table columns.
5. **Within e2e** — background-tab eval vs eval+screenshot same tab.

## Test Index

| Leaf | Scenario |
|------|----------|
| `cli/tab-id-flag-posts-payload` | `session eval --tab-id` posts `tab_id` in job envelope |
| `cli/tab-id-index-conflict` | Both `--tab-id` and `--tab-index` → exit 1 + conflict message |
| `ext-source/resolve-tab-id-window` | background validates `tab_id` in `entry.windowId` |
| `ext-source/resolve-tab-index-order` | 1-based index over capturable tabs in session window |
| `ext-source/attach-reuse-same-tab` | reuse attach same tab; detach when switching `tab_id` |
| `info/json-tab-index-field` | `--json` includes `tabs[].index` and `job_target.tab_index` |
| `info/human-table-columns` | human output lists both tab rows + Job target idx/tab/reason + Recommended + session-page hint |
| `e2e/eval-tab-id-background` | eval with `tab_id` hits background tab without focus |
| `e2e/eval-then-screenshot-same-tab` | eval then screenshot with same `tab_id` both succeed |

**Leaf count: 9**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-tab-targeting
doctest test ./tests/browser-agent-session-tab-targeting
doctest test --label 'slow && ui-automation' ./tests/browser-agent-session-tab-targeting/e2e/...
doctest test ./tests/browser-agent-active-tab-routing
doctest test ./tests/browser-agent-daemon-phase9
```

Static + CLI + info leaves are **RED** until implementer lands tab targeting.
E2e leaves skip when `playwright-debug` absent; **RED** when tool present but
feature unimplemented.

```go
import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface under test.
const (
	ModeCLI       = "cli"
	ModeExtSource = "ext-source"
	ModeInfo      = "info"
	ModeE2E       = "e2e"
)

// CLIOp — CLI flag probes.
const (
	CLIOpTabIDPostsPayload = "tab-id-flag-posts-payload"
	CLIOpTabIDIndexConflict = "tab-id-index-conflict"
)

// ExtSourceTarget — background.js contract probes.
const (
	ExtSrcResolveTabIDWindow    = "resolve-tab-id-window"
	ExtSrcResolveTabIndexOrder  = "resolve-tab-index-order"
	ExtSrcAttachReuseSameTab    = "attach-reuse-same-tab"
)

// InfoOp — session info output probes.
const (
	InfoOpJSONTabIndexField  = "json-tab-index-field"
	InfoOpHumanTableColumns  = "human-table-columns"
)

// PlaywrightOp — e2e script identifiers.
const (
	PlaywrightOpEvalTabIDBackground      = "eval-tab-id-background"
	PlaywrightOpEvalThenScreenshotSameTab = "eval-then-screenshot-same-tab"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string

	CLIOp           string
	ExtSourceTarget string
	InfoOp          string
	PlaywrightOp    string

	SessionID string
	TabID     int64

	CLIArgs         []string
	CLIEnv          map[string]string
	MaxDispatchWait time.Duration

	ReadyTimeout      time.Duration
	PlaywrightTimeout time.Duration
}

// PlaywrightAssertLine is one JSON stdout line from a playwright script.
type PlaywrightAssertLine struct {
	Assert    string `json:"assert"`
	OK        bool   `json:"ok"`
	SessionID string `json:"session_id,omitempty"`
	Extra     map[string]any
}

// Response holds outcomes for all modes.
type Response struct {
	// CLI
	Stdout           string
	Stderr           string
	ExitCode         int
	CLIErr           string
	DispatchTimedOut bool

	ObservedJobType   string
	ObservedJobParams map[string]any
	ObservedJobRaw    string
	ObservedTabID     int64
	WSJobReceived     bool
	TabIDConflictSeen bool

	// ext-source
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	ErrText      string

	// info JSON fields (parsed from --json stdout)
	StdoutJSON     map[string]any
	TabsJSON       []map[string]any
	JobTargetJSON  map[string]any

	// e2e
	Skipped bool
	SkipMsg string

	BaseURL      string
	Addr         string
	ExtensionDir string

	PlaywrightStdout   string
	PlaywrightStderr   string
	PlaywrightExitCode int
	AssertLines        []PlaywrightAssertLine
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	if req.MaxDispatchWait <= 0 {
		req.MaxDispatchWait = 12 * time.Second
	}

	switch req.Mode {
	case ModeCLI:
		return runCLIMode(t, req)
	case ModeExtSource:
		return runExtSourceMode(t, req)
	case ModeInfo:
		return runInfoMode(t, req)
	case ModeE2E:
		return runE2EMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runCLIMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLIOp == "" {
		t.Fatal("CLIOp must be set by leaf Setup")
	}

	switch req.CLIOp {
	case CLIOpTabIDIndexConflict:
		return runCLIConflict(t, req)
	case CLIOpTabIDPostsPayload:
		return runCLITabIDPostsPayload(t, req)
	default:
		return nil, fmt.Errorf("unknown CLIOp %q", req.CLIOp)
	}
}

func runCLIConflict(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	args := req.CLIArgs
	if len(args) == 0 {
		sid := req.SessionID
		if sid == "" {
			sid = "sess-tab-conflict"
		}
		args = []string{
			"session", "eval",
			"--session-id", sid,
			"--tab-id", "111",
			"--tab-index", "2",
			"1+1",
		}
	}
	cliResp, err := invokeHandleCLI(t, req, args)
	resp := &Response{}
	mergeCLIResponse(resp, cliResp)
	combined := strings.ToLower(resp.Stderr + "\n" + resp.CLIErr)
	resp.TabIDConflictSeen = strings.Contains(combined, "cannot use both --tab-id and --tab-index") ||
		(strings.Contains(combined, "--tab-id") && strings.Contains(combined, "--tab-index") &&
			(strings.Contains(combined, "mutually exclusive") || strings.Contains(combined, "cannot use both")))
	return resp, err
}

func runCLITabIDPostsPayload(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by cli grouping Setup")
	}
	tabID := req.TabID
	if tabID == 0 {
		tabID = 216771574
	}
	sid := req.SessionID
	if sid == "" {
		sid = "sess-tab-id-payload"
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, Addr: srv.Addr}

	created, err := createSessionHTTP(srv.BaseURL, sid)
	if err != nil {
		return resp, err
	}
	sid = created

	ext, err := dialFakeExtension(srv.BaseURL, sid, "1.0.0", []string{"browser-agent"})
	if err != nil {
		return resp, fmt.Errorf("fake extension dial: %w", err)
	}
	defer ext.Close()

	var mu sync.Mutex
	ext.OnJob = func(env wsEnvelope) {
		mu.Lock()
		defer mu.Unlock()
		if resp.WSJobReceived {
			return
		}
		resp.WSJobReceived = true
		resp.ObservedJobType = envelopeJobType(env)
		resp.ObservedJobParams = envelopeJobParams(env)
		b, _ := json.Marshal(env)
		resp.ObservedJobRaw = string(b)
		resp.ObservedTabID = envelopeTabID(env)
	}
	ext.AutoCompleteOK = true
	ext.ResultData = map[string]any{"value": 2, "ok": true}
	go ext.Loop()
	time.Sleep(50 * time.Millisecond)

	args := req.CLIArgs
	if len(args) == 0 {
		args = []string{
			"session", "eval",
			"--session-id", sid,
			"--addr", srv.BaseURL,
			"--tab-id", fmt.Sprintf("%d", tabID),
			"1+1",
		}
	}
	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)

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
	return resp, err
}

func runExtSourceMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ExtSourceTarget == "" {
		t.Fatal("ExtSourceTarget must be set by leaf Setup")
	}
	root := req.ModuleRoot
	resp := &Response{}

	candidates := shellBackgroundCandidates(root)
	path, data, ok := firstExistingFile(candidates)
	resp.FileExists = ok
	if ok {
		resp.FoundPaths = []string{path}
		resp.CombinedText = string(data)
	} else {
		resp.ErrText = "shell background.js not found under Chrome-Ext-Browser-Agent"
	}
	return resp, nil
}

func runInfoMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.InfoOp == "" {
		t.Fatal("InfoOp must be set by leaf Setup")
	}
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by info grouping Setup")
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, Addr: srv.Addr}

	sid := req.SessionID
	if sid == "" {
		sid = "sess-tab-info"
	}
	created, err := createSessionHTTP(srv.BaseURL, sid)
	if err != nil {
		return resp, err
	}
	sid = created
	resp.Addr = srv.Addr

	ext, err := dialFakeExtension(srv.BaseURL, sid, "1.0.0", []string{"browser-agent"})
	if err != nil {
		return resp, fmt.Errorf("fake extension dial: %w", err)
	}
	defer ext.Close()

	infoData := map[string]any{
		"tabs": []map[string]any{
			{
				"index": 1, "id": 111, "tab_id": 111,
				"url": srv.BaseURL + "/go?session=" + sid,
				"title": "Browser Agent Session",
				"active": false, "role": "session_page",
			},
			{
				"index": 2, "id": 222, "tab_id": 222,
				"url": "https://example.com/?tab-target=user",
				"title": "Example Domain",
				"active": true, "role": "user",
			},
		},
		"job_target": map[string]any{
			"tab_id": 222, "tab_index": 2,
			"reason": "active_in_session_window",
		},
		"recommended_cli": "browser-agent session eval --tab-id 222 '...'",
		"version":  "1.0.0",
		"features": []any{"browser-agent"},
	}
	ext.AutoCompleteOK = true
	ext.ResultData = infoData
	go ext.Loop()
	time.Sleep(50 * time.Millisecond)

	args := []string{"session", "info", "--session-id", sid, "--addr", srv.BaseURL}
	if req.InfoOp == InfoOpJSONTabIndexField {
		args = append(args, "--json")
	}

	cliResp, err := invokeHandleCLI(t, req, args)
	mergeCLIResponse(resp, cliResp)
	if req.InfoOp == InfoOpJSONTabIndexField {
		parseInfoJSONStdout(resp)
	}
	return resp, err
}

func runE2EMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	pwBin, err := exec.LookPath("playwright-debug")
	if err != nil {
		resp.Skipped = true
		resp.SkipMsg = "playwright-debug not on PATH; skipping E2E leaf"
		t.Skip(resp.SkipMsg)
	}

	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by e2e grouping Setup")
	}
	if req.PlaywrightOp == "" {
		t.Fatal("PlaywrightOp must be set by leaf Setup")
	}

	scriptPath, err := scriptPathForOp(req.PlaywrightOp)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(scriptPath); err != nil {
		return nil, fmt.Errorf("playwright script %s: %w", scriptPath, err)
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	t.Cleanup(cleanup)
	resp.BaseURL = srv.BaseURL
	resp.Addr = srv.Addr

	sid := req.SessionID
	if sid == "" {
		sid = "sess-tab-e2e"
	}
	if _, err := createSessionHTTP(srv.BaseURL, sid); err != nil {
		return nil, err
	}

	extDir, _, err := browseragent.ExtractEmbeddedExtension(req.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("ExtractEmbeddedExtension: %w", err)
	}
	if !filepath.IsAbs(extDir) {
		extDir, err = filepath.Abs(extDir)
		if err != nil {
			return nil, err
		}
	}
	resp.ExtensionDir = extDir

	runPlaywright(t, req, resp, pwBin, scriptPath, srv.BaseURL, sid)
	return resp, nil
}

func scriptPathForOp(op string) (string, error) {
	switch op {
	case PlaywrightOpEvalTabIDBackground:
		return filepath.Join(DOCTEST_ROOT, "e2e", "eval-tab-id-background", "testdata", "eval-tab-id-background.js"), nil
	case PlaywrightOpEvalThenScreenshotSameTab:
		return filepath.Join(DOCTEST_ROOT, "e2e", "eval-then-screenshot-same-tab", "testdata", "eval-screenshot.js"), nil
	default:
		return "", fmt.Errorf("no script mapping for PlaywrightOp %q", op)
	}
}

func runPlaywright(t *testing.T, req *Request, resp *Response, pwBin, scriptPath, baseURL string, sessionIDs ...string) {
	t.Helper()
	timeout := req.PlaywrightTimeout
	if timeout <= 0 {
		timeout = 90 * time.Second
	}

	args := []string{
		"--extension", resp.ExtensionDir,
		"--headed",
		"run", scriptPath,
		baseURL,
	}
	args = append(args, sessionIDs...)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, pwBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	resp.PlaywrightStdout = stdout.String()
	resp.PlaywrightStderr = stderr.String()
	if cmd.ProcessState != nil {
		resp.PlaywrightExitCode = cmd.ProcessState.ExitCode()
	}
	resp.AssertLines = parsePlaywrightAssertLines(resp.PlaywrightStdout)

	if runErr != nil {
		if resp.PlaywrightExitCode == 0 {
			resp.PlaywrightExitCode = 1
		}
	}
}

func parsePlaywrightAssertLines(stdout string) []PlaywrightAssertLine {
	var lines []PlaywrightAssertLine
	sc := bufio.NewScanner(strings.NewReader(stdout))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] != '{' {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		if _, ok := raw["assert"]; !ok {
			continue
		}
		al := PlaywrightAssertLine{Extra: make(map[string]any)}
		if v, ok := raw["assert"].(string); ok {
			al.Assert = v
		}
		if v, ok := raw["ok"].(bool); ok {
			al.OK = v
		}
		if v, ok := raw["session_id"].(string); ok {
			al.SessionID = v
		}
		for k, v := range raw {
			if k == "assert" || k == "ok" || k == "session_id" {
				continue
			}
			al.Extra[k] = v
		}
		lines = append(lines, al)
	}
	return lines
}

func assertLineOK(t *testing.T, lines []PlaywrightAssertLine, name string) {
	t.Helper()
	for _, l := range lines {
		if l.Assert == name {
			if l.OK {
				return
			}
			t.Fatalf("playwright assert %q ok=false session_id=%q extra=%v", name, l.SessionID, l.Extra)
		}
	}
	t.Fatalf("playwright assert %q not found in stdout lines: %v", name, lines)
}

func parseInfoJSONStdout(resp *Response) {
	if resp == nil || strings.TrimSpace(resp.Stdout) == "" {
		return
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Stdout)), &raw); err != nil {
		return
	}
	resp.StdoutJSON = raw

	// session info --json may surface tab fields at top level or under browser{}.
	src := raw
	if browser, ok := raw["browser"].(map[string]any); ok && browser != nil {
		src = browser
	}

	if jt, ok := src["job_target"].(map[string]any); ok {
		resp.JobTargetJSON = jt
	} else if jt, ok := raw["job_target"].(map[string]any); ok {
		resp.JobTargetJSON = jt
	}

	tabsAny, ok := src["tabs"].([]any)
	if !ok {
		tabsAny, _ = raw["tabs"].([]any)
	}
	for _, item := range tabsAny {
		if m, ok := item.(map[string]any); ok {
			resp.TabsJSON = append(resp.TabsJSON, m)
		}
	}
}

func mergeCLIResponse(resp *Response, cli *Response) {
	if cli == nil {
		return
	}
	resp.Stdout = cli.Stdout
	resp.Stderr = cli.Stderr
	resp.ExitCode = cli.ExitCode
	resp.CLIErr = cli.CLIErr
	resp.DispatchTimedOut = cli.DispatchTimedOut
}

func invokeHandleCLI(t *testing.T, req *Request, args []string) (*Response, error) {
	t.Helper()
	maxWait := req.MaxDispatchWait
	var stdout, stderr bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- browseragent.HandleCLI(args, req.CLIEnv, &stdout, &stderr)
	}()

	resp := &Response{}
	select {
	case err := <-done:
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		if err != nil {
			resp.CLIErr = err.Error()
			resp.ExitCode = 1
		} else {
			resp.ExitCode = 0
		}
		return resp, nil
	case <-time.After(maxWait):
		resp.DispatchTimedOut = true
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.ExitCode = 1
		return resp, fmt.Errorf("HandleCLI timed out after %v: args=%v", maxWait, args)
	}
}

func envelopeJobType(env wsEnvelope) string {
	if env.Payload == nil {
		return ""
	}
	if t, ok := env.Payload["type"].(string); ok && t != "" {
		return t
	}
	if t, ok := env.Payload["job_type"].(string); ok && t != "" {
		return t
	}
	if job, ok := env.Payload["job"].(map[string]any); ok {
		if t, ok := job["type"].(string); ok {
			return t
		}
	}
	return ""
}

func envelopeJobParams(env wsEnvelope) map[string]any {
	if env.Payload == nil {
		return nil
	}
	if p, ok := env.Payload["params"].(map[string]any); ok {
		return p
	}
	if job, ok := env.Payload["job"].(map[string]any); ok {
		if p, ok := job["params"].(map[string]any); ok {
			return p
		}
	}
	return nil
}

func envelopeTabID(env wsEnvelope) int64 {
	if env.Payload == nil {
		return 0
	}
	if v, ok := env.Payload["tab_id"]; ok {
		return jsonNumberToInt64(v)
	}
	if job, ok := env.Payload["job"].(map[string]any); ok {
		if v, ok := job["tab_id"]; ok {
			return jsonNumberToInt64(v)
		}
	}
	return 0
}

func jsonNumberToInt64(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}

func shellBackgroundCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "background.js"),
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

// --- RunDaemon harness ---

type daemonServer struct {
	BaseURL string
	Addr    string
	cancel  context.CancelFunc
	done    <-chan error
}

func startDaemonServer(t *testing.T, req *Request) (*daemonServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	ready := req.ReadyTimeout
	if ready <= 0 {
		ready = 10 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	cfg := browseragent.DaemonConfig{
		Addr:    addr,
		BaseDir: req.BaseDir,
		Stdout:  io.Discard,
		Stderr:  io.Discard,
	}

	done := make(chan error, 1)
	go func() {
		_, err := browseragent.RunDaemon(ctx, cfg)
		done <- err
	}()

	baseURL := "http://" + addr
	if err := waitHealth(baseURL, ready); err != nil {
		cancel()
		<-done
		return nil, nil, fmt.Errorf("RunDaemon never healthy at %s: %w", baseURL, err)
	}

	srv := &daemonServer{
		BaseURL: baseURL,
		Addr:    addr,
		cancel:  cancel,
		done:    done,
	}
	cleanup := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}
	return srv, cleanup, nil
}

func waitHealth(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		if healthOK(baseURL) {
			return nil
		}
		last = fmt.Errorf("health not ok")
		time.Sleep(25 * time.Millisecond)
	}
	return last
}

func healthOK(baseURL string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/health", nil)
	if err != nil {
		return false
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	return res.StatusCode == http.StatusOK
}

func createSessionHTTP(baseURL, sessionID string) (string, error) {
	body := map[string]string{}
	if sessionID != "" {
		body["session_id"] = sessionID
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/sessions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("POST /v1/sessions status=%d body=%s", res.StatusCode, strings.TrimSpace(string(out)))
	}
	var parsed map[string]string
	if err := json.Unmarshal(out, &parsed); err != nil {
		return "", err
	}
	sid := parsed["session_id"]
	if sid == "" {
		return "", fmt.Errorf("POST /v1/sessions missing session_id")
	}
	return sid, nil
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
	ResultData     map[string]any
	OnJob          func(wsEnvelope)
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
	_ = f.SendHello()
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
			if f.OnJob != nil {
				f.OnJob(env)
			}
			if f.AutoCompleteOK {
				_ = f.sendResult(env, true, "", f.ResultData)
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
	if data == nil {
		data = map[string]any{"ok": true}
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

var _ = sync.Mutex{}
```