# browser-agent E2E via playwright-debug

Real-browser E2E tests using `playwright-debug` with `--extension <dir>`.
Complements injectable-hook doctests (e.g. `browser-agent-daemon-phase8`).

Depends on daemon phases 1–9, `POST /v1/sessions` without opening Chrome, and
per-session extension WebSocket wiring.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Host** (`RunDaemon`) binds loopback, serves `/v1/health`, `/v1/sessions`,
`/v1/session`, and `/go` session pages. Registry holds zero or more sessions
created via `POST /v1/sessions` — **no** `openChrome` in this tree.

**Embedded Extension** is extracted under `{BaseDir}/extension/…` via
`ExtractEmbeddedExtension`; absolute path is passed to `playwright-debug
--extension`.

**Playwright Harness** skips the leaf when `playwright-debug` is absent from
PATH. Otherwise it launches Chromium headless with the unpacked extension,
runs a leaf `testdata/*.js` script, and collects stdout JSON assert lines
`{"assert":…,"ok":…}`.

**Extension** (MV3 service worker) connects to the daemon per session when a
tab navigates to `/go?session=<id>`.

**Session Page** renders `[data-browser-agent-session-warning]` with
`data-session-id` matching the active session.

## Decision Tree

```
browser-agent-e2e-playwright
├── single-session/                    [one session, extension WS]
│   └── extension-connects/              poll /v1/session until connected
├── session-page/                      [session page DOM]
│   └── warning-banner-visible/          warning banner + session id attr
└── multi-session/                     [two sessions, two tabs]
    └── two-windows-isolated/            both sessions connect independently
```

### Parameter significance (high → low)

1. **Session count** — single vs multi-tab isolation.
2. **Assertion surface** — extension JSON poll vs session-page DOM.
3. **Script** — leaf-specific `testdata/*.js`.

## Test Index

| Leaf | Scenario |
|------|----------|
| `single-session/extension-connects` | Extension connects after `/go`; `extension.connected` true |
| `session-page/warning-banner-visible` | Warning banner present; `data-session-id` matches |
| `multi-session/two-windows-isolated` | Two sessions in two tabs; both connect |

**Leaf count: 3**

## How to Run

```sh
doctest vet ./tests/browser-agent-e2e-playwright
doctest test --label 'slow && ui-automation' ./tests/browser-agent-e2e-playwright
doctest test ./tests/browser-agent-daemon-phase8
```

Requires `playwright-debug` on PATH and Chromium for Playwright. Leaves are
**RED** until harness + extension WS fully land; skip when tool missing.

## Orchestration (Run)

1. `t.TempDir()` → `BaseDir` (root Setup)
2. `RunDaemon` goroutine on `127.0.0.1:0`, capture `baseURL`
3. `POST /v1/sessions` — **no Chrome**
4. `ExtractEmbeddedExtension(BaseDir)` → absolute `ExtensionDir`
5. Skip entire leaf if `playwright-debug` not on PATH
6. `playwright-debug --extension <extDir> --headless run <script.js> <baseURL> <sessionId> [sessionIdB]`
7. Parse stdout JSON lines `{"assert":…,"ok":…}`
8. `t.Cleanup` cancel daemon ctx

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

// PlaywrightOp — leaf scenario identifiers.
const (
	PlaywrightOpExtensionConnects   = "extension-connects"
	PlaywrightOpWarningBanner       = "warning-banner"
	PlaywrightOpTwoWindowsIsolated  = "two-windows-isolated"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	ModuleRoot string
	BaseDir    string

	PlaywrightOp string
	SessionID    string
	SessionIDB   string

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

// Response holds daemon + playwright outcomes.
type Response struct {
	Skipped bool
	SkipMsg string

	BaseURL     string
	Addr        string
	ExtensionDir string

	SessionID  string
	SessionIDB string

	PlaywrightStdout   string
	PlaywrightStderr   string
	PlaywrightExitCode int
	AssertLines        []PlaywrightAssertLine
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	pwBin, err := exec.LookPath("playwright-debug")
	if err != nil {
		resp.Skipped = true
		resp.SkipMsg = "playwright-debug not on PATH; skipping E2E leaf"
		t.Skip(resp.SkipMsg)
	}

	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by root Setup")
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
	case PlaywrightOpExtensionConnects, PlaywrightOpWarningBanner:
		sid := req.SessionID
		if sid == "" {
			t.Fatal("SessionID must be set for single-session leaves")
		}
		if err := createSessionHTTP(srv.BaseURL, sid); err != nil {
			return nil, err
		}
		resp.SessionID = sid
		runPlaywright(t, req, resp, pwBin, scriptPath, srv.BaseURL, sid)
	case PlaywrightOpTwoWindowsIsolated:
		sidA := req.SessionID
		sidB := req.SessionIDB
		if sidA == "" || sidB == "" {
			t.Fatal("SessionID and SessionIDB must be set for two-windows-isolated")
		}
		if err := createSessionHTTP(srv.BaseURL, sidA); err != nil {
			return nil, err
		}
		if err := createSessionHTTP(srv.BaseURL, sidB); err != nil {
			return nil, err
		}
		resp.SessionID = sidA
		resp.SessionIDB = sidB
		runPlaywright(t, req, resp, pwBin, scriptPath, srv.BaseURL, sidA, sidB)
	default:
		return nil, fmt.Errorf("unknown PlaywrightOp %q", req.PlaywrightOp)
	}

	return resp, nil
}

func scriptPathForOp(op string) (string, error) {
	switch op {
	case PlaywrightOpExtensionConnects:
		return filepath.Join(DOCTEST_ROOT, "single-session", "extension-connects", "testdata", "extension-connects.js"), nil
	case PlaywrightOpWarningBanner:
		return filepath.Join(DOCTEST_ROOT, "session-page", "warning-banner-visible", "testdata", "warning-banner.js"), nil
	case PlaywrightOpTwoWindowsIsolated:
		return filepath.Join(DOCTEST_ROOT, "multi-session", "two-windows-isolated", "testdata", "two-windows-isolated.js"), nil
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

	// MV3 extensions do not load a service worker in Chromium classic headless
	// (playwright-debug documents: "extensions may not load"). Use --headed so
	// content script + background WS registration run in real-browser E2E.
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
		// Store outcome; leaf Assert decides pass/fail (RED-friendly).
		resp.PlaywrightExitCode = resp.PlaywrightExitCode
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