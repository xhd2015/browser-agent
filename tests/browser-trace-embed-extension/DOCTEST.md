# browser-trace embed extension — extract, install CLI, launch args, install UX

Exercises **embedding**, **extract**, **install CLI help**, **Chrome launch
args** (no isolated profile), and **session-page install guidance** for the
Chrome-Ext-Capture-API payload shipped inside `browser-trace`.

This tree is **separate** from:

- `./tests/browser-trace/` — sealed capture lifecycle / HAR / deadlines
- `./tests/browser-trace-session-page/` — generic `/v1/session` capability JSON

Those trees must stay green; do not fold these scenarios into them.

**No real Chrome**, no webpack in CI. Tests drive pure Go extract/args/CLI
helpers and short-lived control-server HTTP probes.

## Version

0.0.2

# DSN (Domain Specific Notion)

**User** wants browser-trace to work with the API Capture extension without a
manual multi-step hunt for the extension zip.

**Bundle Tool** (`./script/browser-trace/bundle`) builds the real extension
(or stages a mini fixture) into `browsertrace/embedded/extension/**` for
`//go:embed`.

**browser-trace Binary** embeds that tree. On normal session start and on
`--install-chrome-extension`, the **Extractor** writes it to a stable path:

```text
{BaseDir}/extension/{version}/   # default BaseDir ~/.tmp/browser-trace
  manifest.json
  …
```

yielding absolute **`extension_install_path`** and **`embedded_version`**.

**Install CLI** (`browser-trace --install-chrome-extension`) extracts and
prints path + Developer mode + Load unpacked + `chrome://extensions` help to
**stdout**, ending with trailing `\n`. No full capture session required.

**Chrome Launcher** (best-effort) opens default Chrome in a **new window** to
the session URL with `--load-extension=<extension_install_path>`. It must
**not** pass `--user-data-dir` (no isolated profile). Tests use the pure
**arg builder** — never launch a real browser.

**Control Server** session page surfaces install help when the extension is
not connected or lacks browser-trace support:

- `GET /v1/session` JSON includes `extension_install_path`, `embedded_version`,
  existing `extension.connected` / `supports_browser_trace` / `hint`, and
  install guidance (Load unpacked + path + `chrome://extensions`)
- `GET /go` HTML shows an **install panel** with absolute path,
  `chrome://extensions` as copyable text (chrome: links from http are blocked),
  and stable markers for the panel

**Test Client** in this tree calls package extract/install/args APIs and/or
starts `browsertrace.Run` with `NoOpenChrome` to probe HTTP.

## Decision Tree

```
browser-trace-embed-extension
├── extract/                                   [filesystem extract from embed]
│   ├── first-extract/                           cold extract → dir + manifest + version
│   └── re-extract-same-version/                 second extract → same path, idempotent
├── install-cli/                               [--install-chrome-extension help]
│   └── success-stdout/                          exit 0; path + steps; trailing \n
├── chrome-launch-args/                        [pure arg builder]
│   └── load-extension-no-user-data-dir/         --load-extension=…; no --user-data-dir
├── v1-session-install-help/                   [GET /v1/session install fields]
│   ├── not-connected/                           path + version + install hint present
│   └── connected-supports/                      support OK → install not primary guidance
└── go-install-panel/                          [GET /go HTML install panel]
    └── not-connected-panel/                     path + chrome://extensions + panel markers
```

### Parameter significance (high → low)

1. **Surface / operation** — extract FS vs install CLI vs launch args vs
   `/v1/session` JSON vs `/go` HTML (different contracts and `Run` modes).
2. **Extract pass count** — first extract vs re-extract (idempotency / path stability).
3. **Extension readiness** (HTTP only) — not connected (install help required)
   vs connected + supports (install demoted / not primary).

## Test Index

| Leaf | Scenario (requirement #) |
|------|--------------------------|
| `extract/first-extract` | (#1) Extract embedded extension under BaseDir; manifest + version |
| `extract/re-extract-same-version` | (#2) Re-extract same version → stable path, idempotent |
| `install-cli/success-stdout` | (#3) Install CLI/API: path, chrome://extensions, Load unpacked / Developer; stdout ends `\n` |
| `chrome-launch-args/load-extension-no-user-data-dir` | (#4) Args include `--load-extension=<path>`; omit `--user-data-dir` |
| `v1-session-install-help/not-connected` | (#5) `/v1/session` after extract, no hello: `extension_install_path` + install guidance |
| `v1-session-install-help/connected-supports` | (#7) Connected + supports: install not primary hint (path may still be present) |
| `go-install-panel/not-connected-panel` | (#6) `/go` HTML install panel: absolute path + `chrome://extensions` + markers |

## How to Run

```sh
cd tests/browser-trace-embed-extension
doctest vet .
doctest test -v .
# or from repo root:
doctest vet ./tests/browser-trace-embed-extension
doctest test ./tests/browser-trace-embed-extension
# regression (must stay green):
doctest test ./tests/browser-trace
doctest test ./tests/browser-trace-session-page
```

Requires package `github.com/xhd2015/browser-agent/browsertrace` with
the extract/install/args APIs below (TDD red until implementer lands them).

### Implementer contract (fixture + APIs)

#### Mini-extension / no webpack in CI

- CI **must not** run webpack/npm to execute this tree.
- Ship a **minimal MV3** under `browsertrace/embedded/extension/**` for
  `//go:embed` (see `testdata/mini-extension/` in this tree for the shape:
  `manifest.json` with `"version"`, stub `background.js`).
- Production: `./script/browser-trace/bundle` builds Chrome-Ext-Capture-API and
  stages the real unpacked tree into the same embed dir.
- Optional: allow `Config` / extract helpers to accept an injectable `fs.FS`
  for unit isolation; default is the embedded FS.

#### Runtime layout

```text
{BaseDir}/extension/{version}/manifest.json
```

- `extension_install_path` = absolute path to that directory
- `embedded_version` = `manifest.json` `"version"` string

#### Package APIs (export from `browsertrace`)

```text
ExtractEmbeddedExtension(baseDir string) (installPath, version string, err error)
BuildChromeLaunchArgs(sessionURL, extensionPath string) []string
InstallChromeExtension(w io.Writer, baseDir string) error
```

- `ExtractEmbeddedExtension` is idempotent for the same embedded version
  (stable path; safe to call twice).
- `BuildChromeLaunchArgs` returns argv **without** the Chrome binary name;
  must include `--load-extension=<extensionPath>` (or equivalent joined form)
  and must **not** include any `--user-data-dir` flag/value. Include
  new-window + session URL as product does.
- `InstallChromeExtension` extracts then writes user-facing steps to `w`
  (absolute path, Developer mode, Load unpacked, `chrome://extensions`);
  output **must end with `\n`**. CLI
  `browser-trace --install-chrome-extension` is a thin wrapper: extract+print,
  exit 0, no full capture session.
- Normal `browsertrace.Run` (session start) also extracts so session JSON/HTML
  can expose install path without a prior install-only flag.

#### `GET /v1/session` additions (install-relevant)

When install help is relevant (and ideally always after extract), JSON includes:

```json
{
  "session_id": "...",
  "extension_install_path": "/abs/path/to/extension/1.2.0",
  "embedded_version": "1.2.0",
  "extension": { "connected": false, "supports_browser_trace": false, "...": "..." },
  "hint": "... Load unpacked ... chrome://extensions ..."
}
```

Field names are **top-level** `extension_install_path` and `embedded_version`
(alongside existing `extension` object / `hint`).

#### `GET /go` install panel (not connected / !supports)

- Visible install help with absolute path text
- `chrome://extensions` as **text** (copy affordance OK; no reliance on
  working `<a href="chrome://…">` from http origin)
- Stable markers for tests, e.g. `data-browser-trace-install` and/or
  `id="browser-trace-install"` on the install panel root

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
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browsertrace"
)

// Mode values — surface under test (set by grouping SETUP).
const (
	ModeExtract     = "extract"
	ModeInstallCLI  = "install-cli"
	ModeChromeArgs  = "chrome-args"
	ModeV1Session   = "v1-session"
	ModeGoHTML      = "go"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects which operation Run executes.
	Mode string

	// BaseDir is the browser-trace parent directory (temp per leaf).
	// Extract target: {BaseDir}/extension/{version}/
	BaseDir string

	// ExtractPasses is how many times to call ExtractEmbeddedExtension.
	// 1 = first extract; 2 = re-extract idempotency leaf.
	ExtractPasses int

	// SessionURL is the target URL for chrome arg builder (chrome-args mode).
	SessionURL string
	// ExtensionPath is optional override for chrome-args; empty → extract first.
	ExtensionPath string

	// HTTP session probe fields (v1-session / go modes).
	Addr            string
	SessionSuffix   string
	ReadyTimeout    time.Duration
	CompleteTimeout time.Duration
	NoOpenChrome    bool

	// DoHello stages POST /v1/hello before HTTP probe.
	DoHello       bool
	HelloVersion  string
	HelloFeatures []string
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// Extract
	InstallPath string
	Version     string
	// SecondPassInstallPath is set when ExtractPasses >= 2 (should equal InstallPath).
	SecondPassInstallPath string
	SecondPassVersion     string
	// ManifestPath is InstallPath/manifest.json when extract succeeds.
	ManifestPath string

	// Install CLI / shared stdout
	Stdout   string
	Stderr   string
	ExitCode int
	ErrText  string

	// Chrome launch args (chrome-args mode)
	ChromeArgs []string

	// HTTP probe
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string

	// Parsed /v1/session fields
	SessionID              string
	Phase                  string
	ExtensionConnected     bool
	SupportsBrowserTrace   bool
	ExtensionInstallPath   string
	EmbeddedVersion        string
	Hint                   string
	Raw                    map[string]any

	RealSessionID string
	BaseURL       string
	ProbeURL      string
	RunExitCode   int
	RunErrText    string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}

	switch req.Mode {
	case ModeExtract:
		return runExtract(t, req)
	case ModeInstallCLI:
		return runInstallCLI(t, req)
	case ModeChromeArgs:
		return runChromeArgs(t, req)
	case ModeV1Session, ModeGoHTML:
		return runHTTPProbe(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runExtract(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	passes := req.ExtractPasses
	if passes <= 0 {
		passes = 1
	}

	path1, ver1, err := browsertrace.ExtractEmbeddedExtension(req.BaseDir)
	if err != nil {
		return &Response{ExitCode: 1, ErrText: err.Error()}, err
	}
	resp := &Response{
		InstallPath:  path1,
		Version:      ver1,
		ManifestPath: filepath.Join(path1, "manifest.json"),
		ExitCode:     0,
	}

	if passes >= 2 {
		path2, ver2, err2 := browsertrace.ExtractEmbeddedExtension(req.BaseDir)
		if err2 != nil {
			resp.ExitCode = 1
			resp.ErrText = err2.Error()
			return resp, err2
		}
		resp.SecondPassInstallPath = path2
		resp.SecondPassVersion = ver2
	}
	return resp, nil
}

func runInstallCLI(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	// Package API behind CLI flag --install-chrome-extension.
	err := browsertrace.InstallChromeExtension(&stdout, req.BaseDir)
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		resp.ExitCode = 1
		resp.ErrText = err.Error()
		return resp, err
	}
	resp.ExitCode = 0
	// Best-effort: also populate path/version if extract is queryable.
	if p, v, e := browsertrace.ExtractEmbeddedExtension(req.BaseDir); e == nil {
		resp.InstallPath = p
		resp.Version = v
		resp.ManifestPath = filepath.Join(p, "manifest.json")
	}
	return resp, nil
}

func runChromeArgs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	extPath := req.ExtensionPath
	if extPath == "" {
		p, _, err := browsertrace.ExtractEmbeddedExtension(req.BaseDir)
		if err != nil {
			return &Response{ExitCode: 1, ErrText: err.Error()}, err
		}
		extPath = p
	}
	sessionURL := req.SessionURL
	if sessionURL == "" {
		sessionURL = "http://127.0.0.1:43759/go?session=test-sess"
	}
	args := browsertrace.BuildChromeLaunchArgs(sessionURL, extPath)
	return &Response{
		InstallPath: extPath,
		ChromeArgs:  args,
		ExitCode:    0,
	}, nil
}

func runHTTPProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionSuffix == "" {
		t.Fatal("SessionSuffix must be set for HTTP modes")
	}
	req.NoOpenChrome = true
	if req.ReadyTimeout <= 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout <= 0 {
		req.CompleteTimeout = 2 * time.Second
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
	realSessionID := req.SessionSuffix

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()

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

	if req.DoHello {
		if err := postHello(baseURL, req.HelloVersion, req.HelloFeatures); err != nil {
			runCancel()
			<-runDone
			return nil, fmt.Errorf("POST /v1/hello: %w", err)
		}
		time.Sleep(30 * time.Millisecond)
	}

	var probeURL string
	switch req.Mode {
	case ModeV1Session:
		probeURL = baseURL + "/v1/session?session=" + realSessionID
	case ModeGoHTML:
		probeURL = baseURL + "/go?session=" + realSessionID
	default:
		runCancel()
		<-runDone
		return nil, fmt.Errorf("HTTP mode expected, got %q", req.Mode)
	}

	httpResp, body, err := doGET(probeURL)
	runCancel()
	out := <-runDone

	resp := &Response{
		RealSessionID: realSessionID,
		BaseURL:       baseURL,
		ProbeURL:      probeURL,
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		RunExitCode:   -1,
	}
	if out.result != nil {
		resp.RunExitCode = out.result.ExitCode
	} else if out.err != nil {
		resp.RunExitCode = 1
	}
	if out.err != nil {
		resp.RunErrText = out.err.Error()
	}
	if err != nil {
		return resp, err
	}
	resp.StatusCode = httpResp.StatusCode
	resp.ContentType = httpResp.Header.Get("Content-Type")
	resp.Body = body
	resp.BodyString = string(body)

	if req.Mode == ModeV1Session && len(body) > 0 {
		parseSessionJSON(resp, body)
	}
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

func postHello(baseURL, version string, features []string) error {
	payload := map[string]any{
		"version":  version,
		"features": features,
	}
	if features == nil {
		payload["features"] = []string{}
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	res, err := http.Post(baseURL+"/v1/hello", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	io.Copy(io.Discard, res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("hello status %d", res.StatusCode)
	}
	return nil
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
	return res, body, nil
}

func parseSessionJSON(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.Raw = raw
	resp.SessionID, _ = raw["session_id"].(string)
	resp.Phase, _ = raw["phase"].(string)
	resp.Hint, _ = raw["hint"].(string)
	resp.ExtensionInstallPath, _ = raw["extension_install_path"].(string)
	resp.EmbeddedVersion, _ = raw["embedded_version"].(string)

	if ext, ok := raw["extension"].(map[string]any); ok {
		resp.ExtensionConnected, _ = ext["connected"].(bool)
		resp.SupportsBrowserTrace, _ = ext["supports_browser_trace"].(bool)
	}
}

// Ensure unused import of os is available to descendants via same package gen.
var _ = os.PathSeparator
var _ = strings.TrimSpace
```
