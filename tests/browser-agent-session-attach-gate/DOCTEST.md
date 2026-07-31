# browser-agent session attach gate + multi-tab attach (policy B)

Classic TDD for the **session attach gate** and **policy B multi-tab attach set**:
`chrome.debugger` may attach only while at least one open session control tab
(`…/go?session=<id>`) remains in the **same Chrome window** as the target.
A session holds a **set** of simultaneous attaches; attaching tab B while A is
attached **keeps A** (no sticky switch-detach). Detach **every** tab in the set
when the last session page leaves. Multi-session-tab recount so closing one of
two control tabs does not disarm the session.

| Surface | What is under test |
|---------|-------------------|
| Extension source | Multi-attach keep peers; same-tab reuse + lock; detach-all on leave; attach gate; multi-tab recount |
| E2E (playwright-debug) | Eval succeeds with session open; fails after last page close/navigate-away; stays armed with two pages minus one |

**No real Chrome** for `ext-source` leaves. E2e uses `playwright-debug --extension`
+ embedded extension (same harness as `browser-agent-active-tab-routing` /
`browser-agent-session-tab-targeting` / `browser-agent-e2e-playwright`).

**Policy note:** sticky single `attachedTabId` + switch-detach (policy A) is
**obsolete**. This tree encodes **policy B** only.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Session Control Tab** — tab on the control host at `/go?session=<sessionId>` that
registers the extension WebSocket (`sessions` map: `{ ws, tabId, windowId, … }`).

**Session Attach State** — per-session debugger hold: a **set of attached tab IDs**
(`attachedTabIds` / equivalent), not a single sticky id. Peers stay attached across
jobs until session leave or explicit release.

**Multi-tab Attach (policy B)** — `attachDebuggerForSession(sessionId, tabB)` while
tab A is already in the set **keeps A attached** and adds B. Same-tab second attach
**reuses** (`attachedTabs.has` / set `.has`) without re-attach error. Attach work is
**serialized** per session (`attachLock` or equivalent).

**Attach Gate** — before `chrome.debugger.attach`, require ≥1 open control tab for
the session in the same window as the job target; otherwise refuse with a clear
error (no attach).

**Leave / Recount** — on control-tab close or navigate-away, **re-query** open
`/go?session=<id>` tabs (do not trust `entry.tabId` alone). If remaining count is
0 → **detach every tab** in the session’s attach set and tear down session
registration; if count ≥ 1 → stay armed (rebind `entry.tabId` if needed).

**Background Worker** owns attach/detach via `withDebuggerForSession` /
`attachDebuggerForSession` / `detachDebugger` / `detachSessionDebugger` (or
equivalent helpers).

**Daemon Host** (`RunDaemon`) binds loopback, serves `/v1/sessions`, `/v1/jobs`,
`/v1/session`, `/go`. **Test Client** creates sessions via `POST /v1/sessions`
(no `openChrome`).

**Playwright Harness** (e2e) extracts the embedded extension, launches Chromium
with `playwright-debug --extension --headed`, runs leaf `testdata/*.js`, and
parses stdout JSON assert lines.

```text
Session open in window W
  -> job eval on user tab A in W
  -> attach A; set = {A}

Job eval on tab B in W (session still open)
  -> attach B; set = {A,B}  (A stays attached — no switch-detach)

Last /go?session=S leaves W (close or navigate away)
  -> recount remaining control tabs == 0
  -> detach every tab in set  (banner gone)
  -> subsequent attach/job refused

Two control tabs for S; close one
  -> recount remaining >= 1
  -> stay armed; eval still succeeds
```

## Decision Tree

```
browser-agent-session-attach-gate
├── ext-source/                              [static contract on background.js]
│   ├── detach-on-session-leave/               leave detaches ALL tabs in session attach set
│   ├── attach-gate-requires-session-page/     attach refused without open control tab
│   ├── multi-session-tab-recount/             leave re-queries; not only entry.tabId
│   ├── multi-attach-keep-peers/               attach B keeps A; set state; no switch-detach
│   └── attach-reuse-while-session-open/       same-tab reuse + per-session attachLock
└── e2e/                                     [playwright-debug real browser]
    ├── session-open-eval-succeeds/            control tab open → eval ok
    ├── close-last-session-page-eval-fails/    close last control tab → eval fails
    ├── navigate-away-last-session-page-eval-fails/  navigate away last → eval fails
    └── two-session-pages-close-one-stays-armed/     two control tabs; close one → still eval ok
```

### Parameter significance (high → low)

1. **Test surface** — static extension source vs real-browser E2E.
2. **Within ext-source** — leave detach-all → attach gate → multi-tab recount → multi-attach set → same-tab reuse/lock.
3. **Within e2e** — armed happy path → last-page close → last-page navigate-away → multi-page partial close.

## Test Index

| Leaf | Scenario |
|------|----------|
| `ext-source/detach-on-session-leave` | Session leave / unregister detaches **every** tab in the session attach set |
| `ext-source/attach-gate-requires-session-page` | Attach path gates on open same-window session control tab |
| `ext-source/multi-session-tab-recount` | Leave handler re-queries remaining `/go?session=` tabs (not only `entry.tabId`) |
| `ext-source/multi-attach-keep-peers` | Attach different tab keeps peers; per-session set; **no** switch-detach |
| `ext-source/attach-reuse-while-session-open` | Same-tab reuse (`attachedTabs.has` / set `.has`) + serialize attach (`attachLock`) |
| `e2e/session-open-eval-succeeds` | Session open + user tab → eval job succeeds |
| `e2e/close-last-session-page-eval-fails` | After last control tab closed, subsequent eval does not succeed |
| `e2e/navigate-away-last-session-page-eval-fails` | After last control tab navigates away, subsequent eval does not succeed |
| `e2e/two-session-pages-close-one-stays-armed` | Two control tabs; close one; eval on user tab still succeeds |

**Leaf count: 9**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-attach-gate
doctest test ./tests/browser-agent-session-attach-gate
# L3 (labeled):
doctest test --label 'e2e' ./tests/browser-agent-session-attach-gate
# or: doctest test --label 'slow && ui-automation' ./tests/browser-agent-session-attach-gate
```

Static `ext-source` multi-attach / detach-all leaves are **RED** under sticky
single-`attachedTabId` switch-detach code. Attach gate, multi-tab recount, and
same-tab reuse+lock may already be **GREEN**. E2e leaves skip when
`playwright-debug` is absent; when present, expect implementation-sensitive
results on multi-tab partial close and gate paths.

```go
import (
	"github.com/xhd2015/doctest/session"
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
	ExtSrcDetachOnSessionLeave          = "detach-on-session-leave"
	ExtSrcAttachGateRequiresSessionPage = "attach-gate-requires-session-page"
	ExtSrcMultiSessionTabRecount        = "multi-session-tab-recount"
	ExtSrcMultiAttachKeepPeers          = "multi-attach-keep-peers"
	ExtSrcAttachReuseWhileSessionOpen   = "attach-reuse-while-session-open"
)

// PlaywrightOp for ModeE2E.
const (
	PlaywrightOpSessionOpenEvalSucceeds              = "session-open-eval-succeeds"
	PlaywrightOpCloseLastSessionPageEvalFails        = "close-last-session-page-eval-fails"
	PlaywrightOpNavigateAwayLastSessionPageEvalFails = "navigate-away-last-session-page-eval-fails"
	PlaywrightOpTwoSessionPagesCloseOneStaysArmed    = "two-session-pages-close-one-stays-armed"
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

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}

	switch req.Mode {
	case ModeExtSource:
		return runExtSource(t, req)
	case ModeE2E:
		return runE2E(t, d, req)
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
}

func runE2E(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
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

	scriptPath, err := scriptPathForOp(d, req.PlaywrightOp)
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

	sid := req.SessionID
	if sid == "" {
		t.Fatal("SessionID must be set for e2e leaves")
	}
	if err := createSessionHTTP(srv.BaseURL, sid); err != nil {
		return nil, err
	}
	resp.SessionID = sid
	runPlaywright(t, req, resp, pwBin, scriptPath, srv.BaseURL, sid)
	return resp, nil
}

func scriptPathForOp(d *session.Doctest, op string) (string, error) {
	switch op {
	case PlaywrightOpSessionOpenEvalSucceeds:
		return filepath.Join(d.DOCTEST_ROOT, "e2e", "session-open-eval-succeeds", "testdata", "session-open-eval.js"), nil
	case PlaywrightOpCloseLastSessionPageEvalFails:
		return filepath.Join(d.DOCTEST_ROOT, "e2e", "close-last-session-page-eval-fails", "testdata", "close-last-session-page.js"), nil
	case PlaywrightOpNavigateAwayLastSessionPageEvalFails:
		return filepath.Join(d.DOCTEST_ROOT, "e2e", "navigate-away-last-session-page-eval-fails", "testdata", "navigate-away-last-session-page.js"), nil
	case PlaywrightOpTwoSessionPagesCloseOneStaysArmed:
		return filepath.Join(d.DOCTEST_ROOT, "e2e", "two-session-pages-close-one-stays-armed", "testdata", "two-session-pages-close-one.js"), nil
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
	// Allow long JSON lines from rich extra fields.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
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
