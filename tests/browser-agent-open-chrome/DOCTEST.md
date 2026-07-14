# browser-agent open-chrome — managed Chrome profile

Exercises **managed Chrome** for `browser-agent`: isolated profile under
`~/.browser-agent/managed-chrome/`, embedded extension sync, Chrome argv builder,
injectable launch, CLI `open-chrome`, and **session new** integration.

**No real Chrome.** Tests use pure package APIs and injectable `LaunchFn` /
`ManagedChromeTestHooks` — never spawn a browser process.

Depends on: daemon phases 1–8 (`SessionNew`, `EnsureDaemon`), embedded extension
(`ExtractEmbeddedExtension`).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** runs `browser-agent open-chrome [url]` to open a **managed Chrome
profile** with the embedded extension pre-loaded. Omitting the URL opens a blank
window (no navigation arg).

**Managed Chrome Layout** resolves filesystem paths:

```text
~/.browser-agent/managed-chrome/          # DefaultManagedChromeRoot (via home)
  data/                                   # --user-data-dir for Chrome
  extensions/{version}/                   # embedded MV3 extract target
```

`DefaultManagedChromeLayout()` resolves home → default root. `LayoutFromRoot(root)`
builds `{Root, DataDir, ExtensionsDir}` for `--root` overrides.

**Extension Sync** (`EnsureManagedExtension`) writes the embedded extension under
`{ExtensionsDir}/{version}/` (manifest.json present). Idempotent for the same
embedded version.

**Chrome Args** (`BuildManagedChromeArgs`) returns argv without the binary:

```text
--user-data-dir=<dataDir>
--load-extension=<extensionPath>
--new-window
[<url>]   # only when url non-empty
```

**Open Managed Chrome** (`OpenManagedChrome`) ensures layout + extension sync,
builds args, invokes injectable `LaunchFn` (production: platform launcher).
Returns `OpenChromeResult` with layout, extension path/version, and args.

**Session New** (`SessionNew`) switches to `OpenManagedChrome({URL: sessionURL})`
instead of legacy `openChrome(sessionURL, extPath)`. `--no-open-chrome` unchanged.
Tests spy on `LaunchFn` argv for `--user-data-dir`.

**CLI** — top-level `browser-agent open-chrome [url]` with `--root <dir>`;
`--help` documents managed profile. Pretty stdout on success (path markers).

**Test Client** calls package APIs and `HandleCLI` with injectable hooks; never
launches real Chrome.

## Decision Tree

```
browser-agent-open-chrome
├── layout/                                    [ManagedChromeLayout paths]
│   ├── default-root/                            home → ~/.browser-agent/managed-chrome
│   └── custom-root/                             LayoutFromRoot(--root override)
├── extension-sync/                            [EnsureManagedExtension]
│   ├── extract-writes-version/                  extensions/{ver}/manifest.json
│   └── idempotent-twice/                        second call same path/version
├── chrome-args/                               [BuildManagedChromeArgs pure]
│   ├── blank-window/                            empty url → no url arg; user-data-dir + load-extension
│   └── with-url/                                url appended to argv
├── open-managed/                              [OpenManagedChrome package API]
│   ├── launch-called/                           LaunchFn records argv once
│   └── stdout-markers/                          HandleCLI open-chrome pretty stdout
├── cli-dispatch/                              [HandleCLI help]
│   └── open-chrome-help/                        --help mentions managed profile + open-chrome
└── session-new-integration/                   [SessionNew → OpenManagedChrome]
    └── uses-managed-chrome/                     LaunchFn argv has --user-data-dir
```

### Parameter significance (high → low)

1. **Surface** — layout vs extension sync vs args vs open vs CLI vs session-new.
2. **Root** — default home layout vs custom `--root`.
3. **URL** — blank window vs explicit navigation URL.
4. **Call count** — first extract vs idempotent second call.

## Test Index

| Leaf | Scenario |
|------|----------|
| `layout/default-root` | `DefaultManagedChromeLayout` → `~/.browser-agent/managed-chrome`, `data/`, `extensions/` |
| `layout/custom-root` | `LayoutFromRoot(temp)` → paths under custom root |
| `extension-sync/extract-writes-version` | `EnsureManagedExtension` → `{ExtensionsDir}/{ver}/manifest.json` |
| `extension-sync/idempotent-twice` | Second `EnsureManagedExtension` → same path + version |
| `chrome-args/blank-window` | No url → no http arg; has `--user-data-dir`, `--load-extension`, `--new-window` |
| `chrome-args/with-url` | Non-empty url present in argv |
| `open-managed/launch-called` | `OpenManagedChrome` + `LaunchFn` → called once; managed argv contract |
| `open-managed/stdout-markers` | `HandleCLI open-chrome` → managed profile markers + trailing `\n` |
| `cli-dispatch/open-chrome-help` | `open-chrome --help` mentions managed profile |
| `session-new-integration/uses-managed-chrome` | `SessionNew` → `LaunchFn` argv includes `--user-data-dir` |

**Leaf count: 10**

## How to Run

```sh
doctest vet ./tests/browser-agent-open-chrome
doctest test ./tests/browser-agent-open-chrome
# regression (must stay green after implement):
doctest test ./tests/browser-agent-daemon-phase8
doctest test ./tests/browser-agent-serve-runtime/chrome-hook/...
```

Module: `github.com/xhd2015/browser-agent`. Package under test: `browseragent`
(+ `browseragent/inject` for `ManagedChromeTestHooks`).

### Implementer contract (authoritative for GREEN)

**managed_chrome.go**

```text
const DefaultManagedChromeRoot = "~/.browser-agent/managed-chrome" // resolve via home

type ManagedChromeLayout struct {
    Root, DataDir, ExtensionsDir string
}

func DefaultManagedChromeLayout() (ManagedChromeLayout, error)
func LayoutFromRoot(root string) ManagedChromeLayout

func EnsureManagedExtension(layout ManagedChromeLayout) (path, version string, err error)
// ExtractEmbeddedExtension under layout.ExtensionsDir → {ExtensionsDir}/{version}/

func BuildManagedChromeArgs(dataDir, extensionPath, url string) []string

type OpenManagedChromeConfig struct {
    Root     string   // optional override
    URL      string   // optional; empty = blank window
    LaunchFn func(args []string) error
}

type OpenChromeResult struct {
    Layout         ManagedChromeLayout
    ExtensionPath  string
    ExtensionVer   string
    ChromeArgs     []string
}

func OpenManagedChrome(cfg OpenManagedChromeConfig) (*OpenChromeResult, error)
```

**CLI** — add top-level `open-chrome [url]`; flags `--root`; update `fullHelp`.
`install-chrome-extension` unchanged.

**SessionNew** — replace `openChrome(sessionURL, extPath)` with
`OpenManagedChrome({URL: sessionURL})`; remove separate `ExtractEmbeddedExtension`
on the chrome path.

**Test hooks** (`browseragent/inject/managedchromehooks.go`):

```text
type ManagedChromeHooks struct {
    LaunchFn func(args []string) error
}
var ManagedChromeTestHooks *ManagedChromeHooks
```

`OpenManagedChrome` and `SessionNew` (managed path) consult
`inj.ManagedChromeTestHooks.LaunchFn` when `cfg.LaunchFn` is nil.

**Stdout** (`open-chrome` success): operator-facing lines mentioning managed
profile / data dir / extension path; ends with trailing `\n`.

```go
import (
	"bytes"
	"context"
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
	ModeLayout                = "layout"
	ModeExtensionSync         = "extension-sync"
	ModeChromeArgs            = "chrome-args"
	ModeOpenManaged           = "open-managed"
	ModeCLIDispatch           = "cli-dispatch"
	ModeSessionNewIntegration = "session-new-integration"
)

// LayoutOp — layout probes.
const (
	LayoutOpDefaultRoot = "default-root"
	LayoutOpCustomRoot  = "custom-root"
)

// ExtensionSyncOp — EnsureManagedExtension probes.
const (
	ExtensionSyncOpExtractWritesVersion = "extract-writes-version"
	ExtensionSyncOpIdempotentTwice      = "idempotent-twice"
)

// ChromeArgsOp — BuildManagedChromeArgs probes.
const (
	ChromeArgsOpBlankWindow = "blank-window"
	ChromeArgsOpWithURL     = "with-url"
)

// OpenManagedOp — OpenManagedChrome / CLI stdout probes.
const (
	OpenManagedOpLaunchCalled  = "launch-called"
	OpenManagedOpStdoutMarkers = "stdout-markers"
)

// CLIDispatchOp — HandleCLI probes.
const (
	CLIDispatchOpOpenChromeHelp = "open-chrome-help"
)

// SessionNewIntegrationOp — SessionNew managed chrome probes.
const (
	SessionNewIntegrationOpUsesManagedChrome = "uses-managed-chrome"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	LayoutOp                string
	ExtensionSyncOp         string
	ChromeArgsOp            string
	OpenManagedOp           string
	CLIDispatchOp           string
	SessionNewIntegrationOp string

	// Custom managed-chrome root (--root / LayoutFromRoot).
	ManagedRoot string

	// URL for chrome-args / open-managed / CLI open-chrome.
	URL string

	SessionID string

	ReadyTimeout time.Duration
}

// Response holds outcomes for all modes.
type Response struct {
	Layout browseragent.ManagedChromeLayout

	ExtensionPath  string
	ExtensionVer   string
	ExtensionPath2 string
	ExtensionVer2  string
	ManifestPath   string

	ChromeArgs []string

	LaunchCallCount int
	LaunchArgs      []string

	Stdout   string
	Stderr   string
	CLIErr   string
	ExitCode int

	OpenResult *browseragent.OpenChromeResult

	SessionNewErr string
	SessionID     string
}

// Run executes the scenario selected by req.Mode and leaf Setup narrowing.
func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		t.Fatal("req is nil")
	}
	switch req.Mode {
	case ModeLayout:
		return runLayoutMode(t, req)
	case ModeExtensionSync:
		return runExtensionSyncMode(t, req)
	case ModeChromeArgs:
		return runChromeArgsMode(t, req)
	case ModeOpenManaged:
		return runOpenManagedMode(t, req)
	case ModeCLIDispatch:
		return runCLIDispatchMode(t, req)
	case ModeSessionNewIntegration:
		return runSessionNewIntegrationMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runLayoutMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}
	switch req.LayoutOp {
	case LayoutOpDefaultRoot:
		layout, err := browseragent.DefaultManagedChromeLayout()
		if err != nil {
			return resp, err
		}
		resp.Layout = layout
		return resp, nil
	case LayoutOpCustomRoot:
		if req.ManagedRoot == "" {
			t.Fatal("ManagedRoot must be set for custom-root")
		}
		resp.Layout = browseragent.LayoutFromRoot(req.ManagedRoot)
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown LayoutOp %q", req.LayoutOp)
	}
}

func runExtensionSyncMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ManagedRoot == "" {
		t.Fatal("ManagedRoot must be set for extension-sync")
	}
	layout := browseragent.LayoutFromRoot(req.ManagedRoot)
	resp := &Response{Layout: layout}

	path1, ver1, err := browseragent.EnsureManagedExtension(layout)
	if err != nil {
		return resp, err
	}
	resp.ExtensionPath = path1
	resp.ExtensionVer = ver1
	resp.ManifestPath = filepath.Join(path1, "manifest.json")

	switch req.ExtensionSyncOp {
	case ExtensionSyncOpExtractWritesVersion:
		return resp, nil
	case ExtensionSyncOpIdempotentTwice:
		path2, ver2, err2 := browseragent.EnsureManagedExtension(layout)
		if err2 != nil {
			return resp, err2
		}
		resp.ExtensionPath2 = path2
		resp.ExtensionVer2 = ver2
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown ExtensionSyncOp %q", req.ExtensionSyncOp)
	}
}

func runChromeArgsMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ManagedRoot == "" {
		t.Fatal("ManagedRoot must be set for chrome-args")
	}
	layout := browseragent.LayoutFromRoot(req.ManagedRoot)
	extPath, _, err := browseragent.EnsureManagedExtension(layout)
	if err != nil {
		return nil, err
	}
	url := ""
	switch req.ChromeArgsOp {
	case ChromeArgsOpBlankWindow:
		url = ""
	case ChromeArgsOpWithURL:
		if req.URL == "" {
			req.URL = "https://example.com/session"
		}
		url = req.URL
	default:
		return nil, fmt.Errorf("unknown ChromeArgsOp %q", req.ChromeArgsOp)
	}
	args := browseragent.BuildManagedChromeArgs(layout.DataDir, extPath, url)
	return &Response{
		Layout:        layout,
		ExtensionPath: extPath,
		ChromeArgs:    args,
	}, nil
}

func runOpenManagedMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.OpenManagedOp == "" {
		t.Fatal("OpenManagedOp must be set")
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

	switch req.OpenManagedOp {
	case OpenManagedOpLaunchCalled:
		cfg := browseragent.OpenManagedChromeConfig{
			Root:     req.ManagedRoot,
			LaunchFn: recordLaunch,
		}
		result, err := browseragent.OpenManagedChrome(cfg)
		resp := &Response{OpenResult: result}
		if result != nil {
			resp.Layout = result.Layout
			resp.ExtensionPath = result.ExtensionPath
			resp.ExtensionVer = result.ExtensionVer
			resp.ChromeArgs = result.ChromeArgs
		}
		if err != nil {
			return resp, err
		}
		mu.Lock()
		resp.LaunchCallCount = launchCount
		resp.LaunchArgs = append([]string(nil), launchArgs...)
		mu.Unlock()
		return resp, nil

	case OpenManagedOpStdoutMarkers:
		hooks := &inj.ManagedChromeHooks{LaunchFn: recordLaunch}
		inj.ManagedChromeTestHooks = hooks
		defer func() { inj.ManagedChromeTestHooks = nil }()

		var stdout, stderr bytes.Buffer
		args := []string{"open-managed-chrome"}
		if req.URL != "" {
			args = append(args, req.URL)
		}
		if req.ManagedRoot != "" {
			args = append(args, "--root", req.ManagedRoot)
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
		mu.Lock()
		resp.LaunchCallCount = launchCount
		resp.LaunchArgs = append([]string(nil), launchArgs...)
		mu.Unlock()
		return resp, cliErr

	default:
		return nil, fmt.Errorf("unknown OpenManagedOp %q", req.OpenManagedOp)
	}
}

func runCLIDispatchMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLIDispatchOp == "" {
		t.Fatal("CLIDispatchOp must be set")
	}
	var stdout, stderr bytes.Buffer
	var args []string
	switch req.CLIDispatchOp {
	case CLIDispatchOpOpenChromeHelp:
		args = []string{"open-managed-chrome", "--help"}
	default:
		return nil, fmt.Errorf("unknown CLIDispatchOp %q", req.CLIDispatchOp)
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

func runSessionNewIntegrationMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionNewIntegrationOp == "" {
		t.Fatal("SessionNewIntegrationOp must be set")
	}
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set")
	}

	var launchCount int
	var launchArgs []string
	var mu sync.Mutex
	hooks := &inj.ManagedChromeHooks{
		LaunchFn: func(args []string) error {
			mu.Lock()
			defer mu.Unlock()
			launchCount++
			launchArgs = append([]string(nil), args...)
			return nil
		},
	}
	inj.ManagedChromeTestHooks = hooks
	defer func() { inj.ManagedChromeTestHooks = nil }()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
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
	t.Cleanup(cancel)

	baseURL := "http://" + addr
	if err := waitHealth(baseURL, req.ReadyTimeout); err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	snCfg := browseragent.SessionNewConfig{
		BaseDir:   req.BaseDir,
		Addr:      addr,
		SessionID: req.SessionID,
		Stdout:    &stdout,
		Stderr:    &stderr,
	}
	snErr := browseragent.SessionNew(snCfg)
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if snErr != nil {
		resp.SessionNewErr = snErr.Error()
	}
	mu.Lock()
	resp.LaunchCallCount = launchCount
	resp.LaunchArgs = append([]string(nil), launchArgs...)
	mu.Unlock()

	if snErr != nil {
		return resp, snErr
	}
	resp.SessionID = req.SessionID
	return resp, nil
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