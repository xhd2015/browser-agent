# browser-agent extension install workflow — system Chrome + manual Load unpacked

Exercises the pivot from **managed-chrome-default** to **manual Load unpacked** in
system Chrome (Chrome 137+ ignores `--load-extension`). Covers canonical extension
extract layout, enriched `session new` stdout, renamed `open-managed-chrome`,
`serve`/`RunDaemon` never opening Chrome, `session info` disconnected hints, and
session snapshot `extension_install_path`.

**Classic TDD — RED against current code** (managed chrome default, `open-chrome`,
serve legacy `--session-id` chrome launch).

**No real Chrome.** Injectable `OpenChromeFn` / `LaunchFn` record argv only.

Depends on: daemon phases 1–8 (`SessionNew`, `RunDaemon`, `POST /v1/sessions`),
embedded extension (`ExtractEmbeddedExtension`).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Canonical Extension Layout** resolves the operator install directory under home:

```text
~/.browser-agent/managed-chrome/
  extensions/
    browser-agent/
      <version>/          manifest.json, background.js, bundle-sum.js, …
  data/                   open-managed-chrome profile only
```

`DefaultExtensionInstallLayout()` resolves home → canonical paths.
`EnsureCanonicalExtension()` writes embedded MV3 to
`…/extensions/browser-agent/{version}/` (idempotent per version).

**Session New** (`SessionNew`) operator flow after pivot:

```text
EnsureDaemon → POST /v1/sessions
EnsureCanonicalExtension() → set snapshot extension_install_path
system openChrome(sessionURL only)   # NO --user-data-dir, NO --load-extension
stdout: Extension block + Chrome 137 note + install-chrome-extension hint
        optional open-managed-chrome line in Next:
--no-open-chrome skips system Chrome open only (path still extracted)
```

**Install Chrome Extension** CLI uses canonical path (not `{baseDir}/extension/`).

**Open Managed Chrome** — renamed from `open-chrome`; managed profile with
`--user-data-dir` + `--load-extension`; stderr warns Chrome 137+ ignores
`--load-extension`. `open-chrome` command removed.

**Serve / RunDaemon** — plain `serve` and `RunDaemon` never launch Chrome;
`--no-open-chrome` removed from serve help (or hidden no-op).

**Session Info** — disconnected Next steps mention `install-chrome-extension` +
canonical load path (not `open-chrome`).

**Test Client** calls package APIs and `HandleCLI` with injectable hooks; never
launches real Chrome.

## Decision Tree

```
browser-agent-extension-install-workflow
├── canonical-path/                         [nested DOCTEST — EnsureCanonicalExtension layout]
│   └── (see tests/browser-agent-extension-install-workflow/canonical-path/DOCTEST.md)
├── install-chrome-extension/               [CLI install-chrome-extension]
│   ├── stdout-path-and-steps/                path contains extensions/browser-agent/
│   └── uses-canonical-not-base-dir/          independent of --base-dir
├── session-new/                            [SessionNew workflow]
│   ├── system-chrome-no-user-data-dir/       LaunchFn: --new-window + URL; NO user-data-dir/load-extension
│   ├── extracts-before-open/                   canonical dir exists after SessionNew
│   ├── stdout-extension-block/               path + install-chrome-extension + Chrome 137
│   ├── stdout-no-managed-hint/               does not suggest opening a managed profile
│   ├── no-open-chrome-skips-launch/          NoOpenChrome → LaunchFn 0; path still extracted
│   └── pretty-output-still-has-recipes/      session info/eval/run markers preserved
├── open-managed-chrome/                    [renamed command]
│   ├── cli-dispatch-ok/                      HandleCLI open-managed-chrome exit 0
│   ├── open-chrome-removed/                  HandleCLI open-chrome → unknown command
│   ├── help-mentions-managed/                --help documents open-managed-chrome
│   ├── launch-has-user-data-dir/             LaunchFn has --user-data-dir (managed)
│   └── stderr-chrome137-warning/             stderr warning about --load-extension ignored
├── serve-no-chrome/                        [serve never opens chrome]
│   ├── run-daemon-no-launch/                 RunDaemon/HandleCLI serve → no LaunchFn/OpenChromeFn
│   └── serve-help-no-open-chrome/            serve --help omits --no-open-chrome
├── session-info/                           [session info hints]
│   └── disconnected-install-hint/            human stdout: install-chrome-extension + path
└── snapshot/                               [GET /v1/session path field]
    └── extension-install-path-canonical/     extension_install_path contains extensions/browser-agent/
```

### Parameter significance (high → low)

1. **Surface** — canonical extract vs session-new vs open-managed vs serve vs info vs snapshot.
2. **Chrome launch path** — system (session new) vs managed (`open-managed-chrome`) vs none (serve).
3. **CLI vs package API** — direct API vs `HandleCLI` dispatch.
4. **Flags** — `--no-open-chrome`, `--base-dir`, custom managed `--root`.
5. **Output channel** — stdout markers vs stderr warning vs HTTP JSON field.

## Test Index

| Leaf | Scenario |
|------|----------|
| `install-chrome-extension/stdout-path-and-steps` | CLI stdout path contains `extensions/browser-agent/` + Load unpacked steps |
| `install-chrome-extension/uses-canonical-not-base-dir` | Custom `--base-dir` does not change canonical install path |
| `session-new/system-chrome-no-user-data-dir` | `LaunchFn` argv: `--new-window` + URL; no `--user-data-dir` / `--load-extension` |
| `session-new/extracts-before-open` | Canonical extension dir exists on disk after `SessionNew` |
| `session-new/stdout-extension-block` | Stdout: canonical path + `install-chrome-extension` + Chrome 137 note |
| `session-new/stdout-no-managed-hint` | Stdout does not suggest `open-managed-chrome` |
| `session-new/no-open-chrome-skips-launch` | `NoOpenChrome` → `LaunchFn` 0; canonical path still extracted |
| `session-new/pretty-output-still-has-recipes` | Stdout still has session info/eval/run recipe lines |
| `open-managed-chrome/cli-dispatch-ok` | `HandleCLI ["open-managed-chrome"]` exit 0 |
| `open-managed-chrome/open-chrome-removed` | `HandleCLI ["open-chrome"]` → unknown command error |
| `open-managed-chrome/help-mentions-managed` | Top-level `--help` documents `open-managed-chrome` |
| `open-managed-chrome/launch-has-user-data-dir` | Managed launch argv includes `--user-data-dir` |
| `open-managed-chrome/stderr-chrome137-warning` | Stderr F1 warning: Chrome 137+ ignores `--load-extension` |
| `serve-no-chrome/run-daemon-no-launch` | Plain `serve` / `RunDaemon` → no `LaunchFn` or `OpenChromeFn` |
| `serve-no-chrome/serve-help-no-open-chrome` | `serve --help` omits `--no-open-chrome` |
| `session-info/disconnected-install-hint` | Disconnected human stdout: `install-chrome-extension` + path; no `open-chrome` |
| `snapshot/extension-install-path-canonical` | `GET /v1/session` `extension_install_path` contains `extensions/browser-agent/` |

**Leaf count: 17** (+ **2** nested under `canonical-path/DOCTEST.md` → **19** total)

## How to Run

```sh
doctest vet ./tests/browser-agent-extension-install-workflow
doctest vet ./tests/browser-agent-extension-install-workflow/canonical-path
doctest test ./tests/browser-agent-extension-install-workflow
doctest test ./tests/browser-agent-extension-install-workflow/canonical-path
```

Sibling trees implementer must keep GREEN after landing this feature:

```sh
doctest test ./tests/browser-agent-open-chrome/...      # repoint to open-managed-chrome
doctest test ./tests/browser-agent-daemon-phase8/...    # session-new system chrome args
doctest test ./tests/browser-agent-serve-runtime/...     # serve-no-chrome alignment
doctest test ./tests/browser-agent-daemon-phase10/...   # help text markers
```

Module: `github.com/xhd2015/browser-agent`. Package under test: `browseragent`
(+ `browseragent/inject` for `SessionNewTestHooks` / `ManagedChromeTestHooks`).

### Implementer contract (authoritative for GREEN)

```text
type ExtensionInstallLayout struct {
    Root, BrowserAgentExtensionsDir string
}

func DefaultExtensionInstallLayout() (ExtensionInstallLayout, error)
func EnsureCanonicalExtension() (path, version string, err error)
// path ends with .../extensions/browser-agent/<version>/

func formatSessionNewOutput(...) // Extension block + Note + open-managed-chrome optional line
func WarnLoadExtensionIgnored(stderr, extPath string)

// SessionNew: system openChrome(sessionURL) — no --user-data-dir, no --load-extension
// POST create sets session snapshot extension_install_path to canonical path
// CLI: open-managed-chrome only; remove open-chrome
// serve / RunDaemon: never open Chrome; drop --no-open-chrome from serve help
// install-chrome-extension: canonical path regardless of --base-dir
// session info disconnected: install-chrome-extension + Load unpacked path
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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

// Mode — top-level API surface.
const (
	ModeInstallChromeExt      = "install-chrome-extension"
	ModeSessionNew            = "session-new"
	ModeOpenManagedChrome     = "open-managed-chrome"
	ModeServeNoChrome         = "serve-no-chrome"
	ModeSessionInfo           = "session-info"
	ModeSnapshot              = "snapshot"
)

// InstallChromeExtOp — install-chrome-extension CLI probes.
const (
	InstallChromeExtOpStdoutPathAndSteps     = "stdout-path-and-steps"
	InstallChromeExtOpUsesCanonicalNotBaseDir = "uses-canonical-not-base-dir"
)

// SessionNewOp — SessionNew workflow probes.
const (
	SessionNewOpSystemChromeNoUserDataDir   = "system-chrome-no-user-data-dir"
	SessionNewOpExtractsBeforeOpen          = "extracts-before-open"
	SessionNewOpStdoutExtensionBlock        = "stdout-extension-block"
	SessionNewOpStdoutNoManagedHint         = "stdout-no-managed-hint"
	SessionNewOpNoOpenChromeSkipsLaunch     = "no-open-chrome-skips-launch"
	SessionNewOpPrettyOutputStillHasRecipes = "pretty-output-still-has-recipes"
)

// OpenManagedChromeOp — renamed open-managed-chrome probes.
const (
	OpenManagedChromeOpCLIDispatchOK        = "cli-dispatch-ok"
	OpenManagedChromeOpOpenChromeRemoved    = "open-chrome-removed"
	OpenManagedChromeOpHelpMentionsManaged  = "help-mentions-managed"
	OpenManagedChromeOpLaunchHasUserDataDir = "launch-has-user-data-dir"
	OpenManagedChromeOpStderrChrome137Warn  = "stderr-chrome137-warning"
)

// ServeNoChromeOp — serve / RunDaemon probes.
const (
	ServeNoChromeOpRunDaemonNoLaunch    = "run-daemon-no-launch"
	ServeNoChromeOpServeHelpNoOpenChrome = "serve-help-no-open-chrome"
)

// SessionInfoOp — session info CLI probes.
const (
	SessionInfoOpDisconnectedInstallHint = "disconnected-install-hint"
)

// SnapshotOp — GET /v1/session probes.
const (
	SnapshotOpExtensionInstallPathCanonical = "extension-install-path-canonical"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string
	TestHome   string // isolated HOME for canonical path leaves

	InstallChromeExtOp      string
	SessionNewOp            string
	OpenManagedChromeOp     string
	ServeNoChromeOp         string
	SessionInfoOp           string
	SnapshotOp              string

	SessionID    string
	NoOpenChrome bool
	ManagedRoot  string
	URL          string

	ReadyTimeout time.Duration
}

// Response holds outcomes for all modes.
type Response struct {
	ExtensionPath   string
	ExtensionVer    string
	ExtensionPath2  string
	ExtensionVer2   string
	ManifestPath    string
	LaunchCallCount int
	LaunchArgs      []string

	OpenChromeCallCount  int
	OpenChromeSessionURL string
	OpenChromeExtPath    string

	Stdout   string
	Stderr   string
	CLIErr   string
	ExitCode int

	SessionNewErr string
	SessionID     string
	SessionURL    string

	HTTPStatus                    int
	SessionJSONExtensionInstallPath string
	BodyString                    string

	BaseURL string
	Addr    string
}

// Run executes the scenario selected by req.Mode and leaf Setup narrowing.
func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		t.Fatal("req is nil")
	}
	switch req.Mode {
	case ModeInstallChromeExt:
		return runInstallChromeExtMode(t, req)
	case ModeSessionNew:
		return runSessionNewMode(t, req)
	case ModeOpenManagedChrome:
		return runOpenManagedChromeMode(t, req)
	case ModeServeNoChrome:
		return runServeNoChromeMode(t, req)
	case ModeSessionInfo:
		return runSessionInfoMode(t, req)
	case ModeSnapshot:
		return runSnapshotMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runInstallChromeExtMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.InstallChromeExtOp == "" {
		t.Fatal("InstallChromeExtOp must be set")
	}
	if req.TestHome != "" {
		t.Setenv("HOME", req.TestHome)
	}

	var stdout, stderr bytes.Buffer
	args := []string{"install-chrome-extension"}
	switch req.InstallChromeExtOp {
	case InstallChromeExtOpStdoutPathAndSteps:
		// default: no --base-dir
	case InstallChromeExtOpUsesCanonicalNotBaseDir:
		if req.BaseDir == "" {
			t.Fatal("BaseDir must be set for uses-canonical-not-base-dir")
		}
		args = append(args, "--base-dir", req.BaseDir)
	default:
		return nil, fmt.Errorf("unknown InstallChromeExtOp %q", req.InstallChromeExtOp)
	}

	cliErr := browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	}
	return resp, cliErr
}

func runSessionNewMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionNewOp == "" {
		t.Fatal("SessionNewOp must be set")
	}
	if req.TestHome != "" {
		t.Setenv("HOME", req.TestHome)
	}
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set")
	}

	var launchCount int
	var launchArgs []string
	var openCount int
	var openURL, openExt string
	var hookMu sync.Mutex

	hooks := &inj.ManagedChromeHooks{
		LaunchFn: func(args []string) error {
			hookMu.Lock()
			defer hookMu.Unlock()
			launchCount++
			launchArgs = append([]string(nil), args...)
			return nil
		},
	}
	inj.ManagedChromeTestHooks = hooks
	defer func() { inj.ManagedChromeTestHooks = nil }()

	// Do not inject OpenChromeFn — production path must exercise system/managed launch
	// via OpenManagedChrome → ManagedChromeTestHooks.LaunchFn for argv assertions.

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	t.Cleanup(cleanup)

	var stdout, stderr bytes.Buffer
	snCfg := browseragent.SessionNewConfig{
		BaseDir:      req.BaseDir,
		Addr:         srv.Addr,
		SessionID:    req.SessionID,
		NoOpenChrome: req.NoOpenChrome,
		Stdout:       &stdout,
		Stderr:       &stderr,
	}
	snErr := browseragent.SessionNew(snCfg)
	resp := &Response{
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		BaseURL:       srv.BaseURL,
		Addr:          srv.Addr,
		SessionID:     req.SessionID,
	}
	if snErr != nil {
		resp.SessionNewErr = snErr.Error()
	}

	hookMu.Lock()
	resp.LaunchCallCount = launchCount
	resp.LaunchArgs = append([]string(nil), launchArgs...)
	resp.OpenChromeCallCount = openCount
	resp.OpenChromeSessionURL = openURL
	resp.OpenChromeExtPath = openExt
	hookMu.Unlock()

	if snErr != nil {
		return resp, snErr
	}

	if path, ver, ok := findCanonicalExtensionUnderHome(req.TestHome); ok {
		resp.ExtensionPath = path
		resp.ExtensionVer = ver
	}

	return resp, nil
}

func runOpenManagedChromeMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.OpenManagedChromeOp == "" {
		t.Fatal("OpenManagedChromeOp must be set")
	}
	if req.TestHome != "" {
		t.Setenv("HOME", req.TestHome)
	}

	var launchCount int
	var launchArgs []string
	var mu sync.Mutex
	recordLaunch := func(args []string) error {
		mu.Lock()
		defer mu.Unlock()
		launchCount++
		launchArgs = append([]string(nil), args...)
		return nil
	}

	inj.ManagedChromeTestHooks = &inj.ManagedChromeHooks{LaunchFn: recordLaunch}
	defer func() { inj.ManagedChromeTestHooks = nil }()

	var stdout, stderr bytes.Buffer
	var args []string
	var cliErr error

	switch req.OpenManagedChromeOp {
	case OpenManagedChromeOpCLIDispatchOK:
		args = []string{"open-managed-chrome"}
		if req.URL != "" {
			args = append(args, req.URL)
		}
		cliErr = browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)

	case OpenManagedChromeOpOpenChromeRemoved:
		args = []string{"open-chrome"}
		cliErr = browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)

	case OpenManagedChromeOpHelpMentionsManaged:
		args = []string{"--help"}
		cliErr = browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)

	case OpenManagedChromeOpLaunchHasUserDataDir, OpenManagedChromeOpStderrChrome137Warn:
		args = []string{"open-managed-chrome"}
		if req.URL == "" {
			req.URL = "https://example.com/session"
		}
		args = append(args, req.URL)
		if req.ManagedRoot != "" {
			args = append(args, "--root", req.ManagedRoot)
		}
		cliErr = browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)

	default:
		return nil, fmt.Errorf("unknown OpenManagedChromeOp %q", req.OpenManagedChromeOp)
	}

	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	}
	mu.Lock()
	resp.LaunchCallCount = launchCount
	resp.LaunchArgs = append([]string(nil), launchArgs...)
	mu.Unlock()
	return resp, cliErr
}

func runServeNoChromeMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ServeNoChromeOp == "" {
		t.Fatal("ServeNoChromeOp must be set")
	}

	var launchCount, openCount int
	var mu sync.Mutex

	inj.ManagedChromeTestHooks = &inj.ManagedChromeHooks{
		LaunchFn: func(args []string) error {
			mu.Lock()
			defer mu.Unlock()
			launchCount++
			return nil
		},
	}
	defer func() { inj.ManagedChromeTestHooks = nil }()

	inj.SessionNewTestHooks = &inj.SessionNewHooks{
		OpenChromeFn: func(sessionURL, extPath string) error {
			mu.Lock()
			defer mu.Unlock()
			openCount++
			return nil
		},
	}
	defer func() { inj.SessionNewTestHooks = nil }()

	var stdout, stderr bytes.Buffer
	var cliErr error

	switch req.ServeNoChromeOp {
	case ServeNoChromeOpRunDaemonNoLaunch:
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}
		addr := ln.Addr().String()
		_ = ln.Close()

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			cfg := browseragent.DaemonConfig{
				Addr:    addr,
				BaseDir: req.BaseDir,
				Stdout:  io.Discard,
				Stderr:  io.Discard,
			}
			_, err := browseragent.RunDaemon(ctx, cfg)
			done <- err
		}()
		t.Cleanup(cancel)

		baseURL := "http://" + addr
		if err := waitHealth(baseURL, req.ReadyTimeout); err != nil {
			cancel()
			<-done
			return nil, err
		}
		cancel()
		serveErr := <-done
		if serveErr != nil && !strings.Contains(serveErr.Error(), "context canceled") {
			return nil, serveErr
		}

	case ServeNoChromeOpServeHelpNoOpenChrome:
		cliErr = browseragent.HandleCLI([]string{"serve", "--help"}, map[string]string{}, &stdout, &stderr)

	default:
		return nil, fmt.Errorf("unknown ServeNoChromeOp %q", req.ServeNoChromeOp)
	}

	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	}
	mu.Lock()
	resp.LaunchCallCount = launchCount
	resp.OpenChromeCallCount = openCount
	mu.Unlock()
	return resp, cliErr
}

func runSessionInfoMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionInfoOp == "" {
		t.Fatal("SessionInfoOp must be set")
	}
	if req.TestHome != "" {
		t.Setenv("HOME", req.TestHome)
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	t.Cleanup(cleanup)

	// Create disconnected session via POST (no extension hello).
	createURL := strings.TrimRight(srv.BaseURL, "/") + "/v1/sessions"
	body := fmt.Sprintf(`{"session_id":%q}`, req.SessionID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, createURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("POST /v1/sessions status %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}

	var stdout, stderr bytes.Buffer
	host, portStr, _ := net.SplitHostPort(srv.Addr)
	args := []string{
		"session", "info",
		"--session-id", req.SessionID,
		"--base-dir", req.BaseDir,
		"--host", host,
		"--server-port", portStr,
	}
	cliErr := browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)
	resp := &Response{
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		SessionID: req.SessionID,
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	}
	return resp, cliErr
}

func runSnapshotMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SnapshotOp == "" {
		t.Fatal("SnapshotOp must be set")
	}
	if req.TestHome != "" {
		t.Setenv("HOME", req.TestHome)
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	t.Cleanup(cleanup)

	var stdout, stderr bytes.Buffer
	snCfg := browseragent.SessionNewConfig{
		BaseDir:   req.BaseDir,
		Addr:      srv.Addr,
		SessionID: req.SessionID,
		NoOpenChrome: true,
		Stdout:    &stdout,
		Stderr:    &stderr,
	}
	if err := browseragent.SessionNew(snCfg); err != nil {
		return &Response{SessionNewErr: err.Error()}, err
	}

	u := strings.TrimRight(srv.BaseURL, "/") + "/v1/session?session=" + req.SessionID
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	resp := &Response{
		HTTPStatus:  res.StatusCode,
		BodyString:  string(body),
		SessionID:   req.SessionID,
		BaseURL:     srv.BaseURL,
		Stdout:      stdout.String(),
	}

	var snap struct {
		ExtensionInstallPath string `json:"extension_install_path"`
	}
	if err := json.Unmarshal(body, &snap); err == nil {
		resp.SessionJSONExtensionInstallPath = snap.ExtensionInstallPath
	}
	return resp, nil
}

type daemonServer struct {
	Addr    string
	BaseURL string
}

func startDaemonServer(t *testing.T, req *Request) (*daemonServer, func(), error) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	addr := ln.Addr().String()
	_ = ln.Close()

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
	if err := waitHealth(baseURL, req.ReadyTimeout); err != nil {
		cancel()
		<-done
		return nil, nil, err
	}

	cleanup := func() {
		cancel()
		<-done
	}
	return &daemonServer{Addr: addr, BaseURL: baseURL}, cleanup, nil
}

func waitHealth(baseURL string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(strings.TrimRight(baseURL, "/") + "/v1/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("health check timed out for %s", baseURL)
}
```
