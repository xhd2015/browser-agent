# browser-agent active tab routing (session window)

Coverage backfill for the **active tab routing** fix in
`pickTargetTabIdForSession`: jobs attach to the **active capturable tab** in the
session window (`entry.windowId`), not only the registered `/go?session=` control
page tab.

| Surface | What is under test |
|---------|-------------------|
| Extension source | `pickTargetTabIdForSession` prefers `active: true` + `windowId`; session-page fallback retained |
| E2E (playwright-debug) | `info` + `eval` jobs run on the user tab when it is active in the session window |

Complements `tests/browser-agent-daemon-phase9/ext-source/job-session-routing` (P4
session-scoped routing) with **window-active-tab priority** and a real-browser
regression script.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Session Window** — Chrome window where `/go?session=<id>` is open. The
extension `sessions` map entry stores `{ ws, tabId, windowId, controlPort, … }`.

**Session Page Tab** — the tab that navigated to `/go?session=<id>` and
registered the extension WebSocket. Historically jobs pinned here even when
another tab in the same window was active.

**Job Target Tab** — tab `chrome.debugger` attaches to for `eval`, `cdp`, and
`screenshot` jobs.

**Background Worker** resolves the job target via `pickTargetTabIdForSession`:

1. Active capturable tab in `entry.windowId`
2. Registered `entry.tabId` (session control page)
3. URL scan for `/go?session=<id>`

**Daemon Host** (`RunDaemon`) binds loopback, serves `/v1/sessions`, `/v1/jobs`,
`/v1/session`, and `/go`. **Test Client** creates sessions via
`POST /v1/sessions` (no `openChrome`).

**Playwright Harness** (e2e leaves) extracts the embedded extension, launches
Chromium with `playwright-debug --extension --headed`, runs leaf
`testdata/*.js`, and parses stdout JSON assert lines.

```text
Tab 1: /go?session=S  -> register -> sessions[S] = {tabId, windowId, ws}
Tab 2: user URL       -> bringToFront (active in session window)

Background on job (session_id = S)
  -> pickTargetTabIdForSession(S)
  -> prefer active tab in entry.windowId
  -> attachDebugger(activeTabId)
  -> eval returns user tab URL (not session page)
```

## Decision Tree

```
browser-agent-active-tab-routing
├── ext-source/                          [static contract on background.js]
│   └── window-active-tab-priority/        active+windowId before session-page fallback
└── e2e/                                 [playwright-debug real browser]
    └── eval-on-active-user-tab/           info+eval on active user tab
```

### Parameter significance (high → low)

1. **Test surface** — static source contract vs real-browser E2E.
2. **Within ext-source** — `pickTargetTabIdForSession` active-tab priority tokens.
3. **Within e2e** — playwright script + JSON assert lines.

## Test Index

| Leaf | Scenario |
|------|----------|
| `ext-source/window-active-tab-priority` | `pickTargetTabIdForSession` queries `active: true` scoped to `entry.windowId`; session-page fallback retained |
| `e2e/eval-on-active-user-tab` | Two tabs in one window; `eval` returns user tab URL with `LOOP_MARKER` |

**Leaf count: 2**

## How to Run

```sh
doctest vet ./tests/browser-agent-active-tab-routing
doctest test ./tests/browser-agent-active-tab-routing
doctest test --label 'slow && ui-automation' ./tests/browser-agent-active-tab-routing/e2e/...
```

E2e leaves skip when `playwright-debug` is absent from PATH. **GREEN** expected
(coverage backfill — fix already landed).

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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface under test.
const (
	ModeExtSource = "ext-source"
	ModeE2E       = "e2e"
)

// ExtSourceTarget for ModeExtSource.
const (
	ExtSrcWindowActiveTabPriority = "window-active-tab-priority"
)

// PlaywrightOp for ModeE2E.
const (
	PlaywrightOpEvalOnActiveUserTab = "eval-on-active-user-tab"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string

	ExtSourceTarget string
	PlaywrightOp    string
	SessionID       string

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

// Response holds ext-source probe or daemon + playwright outcomes.
type Response struct {
	// ext-source
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	ErrText      string

	// e2e
	Skipped bool
	SkipMsg string

	BaseURL      string
	Addr         string
	ExtensionDir string
	SessionID    string

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

	switch req.Mode {
	case ModeExtSource:
		return runExtSource(t, req)
	case ModeE2E:
		return runE2E(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runExtSource(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ExtSourceTarget == "" {
		t.Fatal("ExtSourceTarget must be set")
	}
	root := req.ModuleRoot
	resp := &Response{FileContents: map[string]string{}}

	switch req.ExtSourceTarget {
	case ExtSrcWindowActiveTabPriority:
		candidates := shellBackgroundCandidates(root)
		path, data, ok := firstExistingFile(candidates)
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "shell background.js not found under Chrome-Ext-Browser-Agent"
		}
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown ExtSourceTarget %q", req.ExtSourceTarget)
	}
}

func runE2E(t *testing.T, req *Request) (*Response, error) {
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

	switch req.PlaywrightOp {
	case PlaywrightOpEvalOnActiveUserTab:
		sid := req.SessionID
		if sid == "" {
			t.Fatal("SessionID must be set for eval-on-active-user-tab")
		}
		if err := createSessionHTTP(srv.BaseURL, sid); err != nil {
			return nil, err
		}
		resp.SessionID = sid
		runPlaywright(t, req, resp, pwBin, scriptPath, srv.BaseURL, sid)
	default:
		return nil, fmt.Errorf("unknown PlaywrightOp %q", req.PlaywrightOp)
	}

	return resp, nil
}

func scriptPathForOp(op string) (string, error) {
	switch op {
	case PlaywrightOpEvalOnActiveUserTab:
		return filepath.Join(DOCTEST_ROOT, "e2e", "eval-on-active-user-tab", "testdata", "active-tab-routing.js"), nil
	default:
		return "", fmt.Errorf("no script mapping for PlaywrightOp %q", op)
	}
}

func runPlaywright(t *testing.T, req *Request, resp *Response, pwBin, scriptPath, baseURL string, sessionIDs ...string) {
	t.Helper()
	timeout := req.PlaywrightTimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
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

// --- daemon harness ---

type daemonServer struct {
	BaseURL string
	Addr    string
	cancel  context.CancelFunc
	done    <-chan error
}

func startDaemonServer(t *testing.T, req *Request) (*daemonServer, func(), error) {
	t.Helper()
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

func createSessionHTTP(baseURL, sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	body := fmt.Sprintf(`{"session_id":%q}`, sessionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/sessions", strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("POST /v1/sessions status=%d body=%s", res.StatusCode, strings.TrimSpace(string(out)))
	}
	return nil
}

var _ = sync.Mutex{}
```