# browser-agent CLI + React product shell + extension embed

Exercises the **operator-facing product shell** on top of the sealed
`browseragent` control plane MVP (`tests/browser-agent/` — **do not edit**):

| Surface | What is under test |
|---------|-------------------|
| CLI dispatch | bare / `--help` / `session info`+`session eval` missing session via `HandleCLI` |
| CLI side-commands | `session info` / `session eval` against live in-process serve + fake WS |
| ProductConfig | browser-agent (43761) + browser-trace (43759) dual export |
| SPA embed | GET `/go` or `/` React root + product hooks + install markers |
| Extension extract | embed extract, re-extract, manifest 43761, install CLI stdout |
| Chrome launch args | pure `--load-extension=…`; no `--user-data-dir` |
| React source layout | `react/src/products`, apps entries, `InstallGuideline` on disk |
| Ext shell on disk | `Chrome-Ext-Browser-Agent` manifest name + host **43761** |

**No real Chrome.** **No real agent-run.** **No webpack/npm in CI.** Fake WS
completes eval jobs. Filesystem leaves read module tree via `DOCTEST_ROOT`.

Sealed regressions (must stay green after implement):

```sh
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-trace...
doctest test ./tests/browser-trace-session-page
doctest test ./tests/browser-trace-install-panel
doctest test ./tests/browser-trace-embed-extension
```

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** runs **`browser-agent`** CLI (`cmd/browser-agent` → package
`HandleCLI`) for serve and **nested** side-commands under `session`. **Agent**
(or scripts) call `session info` / `session eval` after session resolve.

**HandleCLI** (package API preferred over binary shell-out):

```text
HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
```

- bare / empty args → brief usage (mentions `serve`); non-nil error; trailing `\n`
- `-h` / `--help` → help listing `serve`, `session`, nested cmds; nil error; trailing `\n`
- `session info` / `session eval` → resolve session (`--session-id` or env
  `BROWSER_AGENT_SESSION_ID`); missing both → error mentions **both** sources
- `session eval <expr>` → POST `/v1/jobs` type=eval; print result + trailing `\n`
- `session info` → session snapshot / info job; stdout JSON-ish + `\n`
- Flat `info`/`eval` are **not** side-command handlers (unknown / brief error)
- `--addr` (or equivalent) points side-commands at control server (tests use free port)

**Control Server** is existing `browseragent.Run` (NoOpenChrome, NoAgentRun).
Session SPA HTML comes from **embedded session-page dist** (Vite React build
staged under `browseragent/embedded/session-page/` or equivalent).

**ProductConfig** (Go + TS mirror) parameterizes ports/names/features:

```text
ProductConfig {
  id: "browser-agent" | "browser-trace"
  displayName, cliName
  controlPort: 43761 | 43759
  features: ["browser-agent", …] | ["browser-trace", …]
  pageMarkerGlobal: "__BROWSER_AGENT_EXT__" | "__BROWSER_TRACE_EXT__"
  extensionDirName: "Chrome-Ext-Browser-Agent" | "Chrome-Ext-Capture-API"
}
```

**Vite React** lives under module `react/` (or `project-api-capture/react/`):

```text
react/src/products/{types,browser-agent,browser-trace}.ts
react/src/ui/InstallGuideline.tsx
react/src/apps/session-page/main.tsx
react/src/apps/popup/main.tsx
```

**Chrome-Ext-Browser-Agent** on disk: MV3 shell with manifest name Browser Agent
and host_permissions / content_scripts targeting port **43761**.

**Extractor** (mirror browser-trace):

```text
{BaseDir}/extension/{version}/
ExtractEmbeddedExtension(baseDir) → installPath, version
InstallChromeExtension(w, baseDir) → path + Load unpacked + chrome://extensions + \n
BuildChromeArgs / BuildChromeLaunchArgs(sessionURL, extPath) → argv without --user-data-dir
```

**Test Client** in this tree:

- CLI leaves call `HandleCLI` with injectable env + buffers (no `go run` binary).
- Side-command leaves start `browseragent.Run` + fake extension WS, then HandleCLI.
- Product / chrome-args / extract call pure package APIs.
- React-src / ext-shell read files under **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.

## Decision Tree

```
browser-agent-cli-react
├── cli-dispatch/                              [HandleCLI only]
│   ├── bare/                                    A1 empty args → brief + err + \n
│   ├── help/                                    A2 --help lists serve,session + nested +\n
│   ├── eval-without-session/                    A3 session eval missing session → both sources
│   └── info-without-session/                    A4 session info missing session → both sources
├── cli-sidecmd/                               [serve + HandleCLI nested side-command]
│   ├── eval-with-fake-extension/                B1 session eval --session-id; fake WS; ok +\n
│   └── info-session-snapshot/                   B2 session info --session-id; session_id +\n
├── product-config/                            [ProductConfig export]
│   ├── browser-agent/                           C1 port 43761, id browser-agent
│   └── browser-trace/                           C2 port 43759, id browser-trace
├── spa-embed/                                 [GET /go HTML product SPA]
│   ├── root-mount-and-hooks/                    D1 root + /v1/session + 43761 + browser-agent
│   └── install-guideline-markers/               D2 chrome://extensions / Load unpacked
├── extension-extract/                         [embed extract + install CLI]
│   ├── first-extract/                           E1 dir + manifest + version
│   ├── re-extract-same-version/                 E2 stable path idempotent
│   ├── manifest-hosts-43761/                    E3 extracted manifest mentions 43761
│   └── install-cli-stdout/                      E4 path + chrome://extensions + \n
├── chrome-launch-args/                        [pure arg builder]
│   └── load-extension-no-user-data-dir/         F1 --load-extension; no user-data-dir
├── react-src/                                 [module filesystem layout]
│   ├── products-browser-agent/                  G1 react/src/products/browser-agent.*
│   ├── apps-entries/                            G2 session-page + popup main entries
│   └── install-guideline-component/             G3 ui/InstallGuideline.*
└── ext-shell/                                 [Chrome-Ext-Browser-Agent on disk]
    └── chrome-ext-browser-agent-manifest/       H1 name Browser Agent; hosts 43761
```

### Parameter significance (high → low)

1. **Surface / Mode** — CLI dispatch vs live side-cmd vs product vs SPA vs extract
   vs chrome-args vs react-src vs ext-shell (different `Run` branches).
2. **Within CLI** — dispatch kind (bare/help/missing-session) vs live success path
   (eval vs info).
3. **Product id** — browser-agent vs browser-trace.
4. **Extract op** — first vs re vs manifest content vs install stdout.
5. **React probe** — products file vs apps entries vs InstallGuideline component.

## Test Index

| Leaf | Scenario |
|------|----------|
| `cli-dispatch/bare` | (A1) `HandleCLI([])` → non-nil err; usage mentions `serve`; trailing `\n` |
| `cli-dispatch/help` | (A2) `--help` → nil err; lists `serve`, `session`, nested `info`/`eval`; trailing `\n` |
| `cli-dispatch/eval-without-session` | (A3) `session eval '1+1'` no session → err mentions `--session-id` + `BROWSER_AGENT_SESSION_ID` |
| `cli-dispatch/info-without-session` | (A4) `session info` no session → same dual mention |
| `cli-sidecmd/eval-with-fake-extension` | (B1) serve + fake WS; `session eval --session-id X --addr … '1+1'` → nil; result; `\n` |
| `cli-sidecmd/info-session-snapshot` | (B2) serve; `session info --session-id X --addr …` → session_id; connected field; `\n` |
| `product-config/browser-agent` | (C1) ProductBrowserAgent: id/cli browser-agent; port 43761; feature browser-agent |
| `product-config/browser-trace` | (C2) ProductBrowserTrace: id browser-trace; port 43759 |
| `spa-embed/root-mount-and-hooks` | (D1) GET `/go` or `/`: root mount + `/v1/session` + `43761` + `browser-agent` |
| `spa-embed/install-guideline-markers` | (D2) not connected: `chrome://extensions` or Load unpacked guidance |
| `extension-extract/first-extract` | (E1) extract → abs path under `{BaseDir}/extension/{version}`; manifest |
| `extension-extract/re-extract-same-version` | (E2) second extract → same path/version |
| `extension-extract/manifest-hosts-43761` | (E3) extracted manifest text mentions **43761** |
| `extension-extract/install-cli-stdout` | (E4) InstallChromeExtension stdout: path + chrome://extensions + load unpacked; `\n` |
| `chrome-launch-args/load-extension-no-user-data-dir` | (F1) `--load-extension=<path>`; no `--user-data-dir` |
| `react-src/products-browser-agent` | (G1) `react/src/products/browser-agent.{ts,tsx,js}` exists; contains `43761` + `browser-agent` |
| `react-src/apps-entries` | (G2) session-page + popup app entry files under `react/src/apps/` |
| `react-src/install-guideline-component` | (G3) `InstallGuideline` under `react/src/ui/` |
| `ext-shell/chrome-ext-browser-agent-manifest` | (H1) `Chrome-Ext-Browser-Agent` manifest: Browser Agent + 43761 |

**Leaf count: 19**

## How to Run

```sh
cd project-api-capture
doctest vet ./tests/browser-agent-cli-react
doctest test ./tests/browser-agent-cli-react/...
# or:
cd tests/browser-agent-cli-react && doctest vet . && doctest test -v .
```

Module: `github.com/xhd2015/browser-agent`.  
Package under test primarily `…/browseragent` (CLI Handle may live in
`browseragent` or `browseragent/cli` re-exported as `HandleCLI`).

### Implementer contract (authoritative for GREEN)

**CLI**

```go
// Prefer exported from browseragent (or re-export from browseragent/cli):
func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
```

- `args` are argv **after** the binary name (no `browser-agent` prefix required).
- `env` is injectable; when key missing, treat as unset (do **not** fall back to
  process env for session id when `env != nil` — tests pass explicit maps).
  When `env == nil`, process env may be used (not required for this tree).
- bare → brief usage on stdout **or** stderr; prefer non-nil error; trailing `\n`.
- `--help` / `-h` → help; **nil** error; no `os.Exit`; trailing `\n`.
- Side-commands need `--session-id` / env and `--addr` (default
  `http://127.0.0.1:43761` or host:port form).
- Successful side-command stdout ends with `\n`.

**ProductConfig**

```go
type ProductConfig struct {
    ID               string   // "browser-agent" | "browser-trace"
    DisplayName      string
    CLIName          string
    ControlPort      int      // 43761 | 43759 (string form also OK via helper)
    Features         []string
    PageMarkerGlobal string
    ExtensionDirName string
}

var ProductBrowserAgent ProductConfig // required
var ProductBrowserTrace ProductConfig // required for C2 dual-export design
```

Also keep existing `DefaultAddr` / `DefaultControlPort` = 43761.

**Embed / extract / args** (mirror browser-trace names OK)

```go
ExtractEmbeddedExtension(baseDir string) (installPath, version string, err error)
InstallChromeExtension(w io.Writer, baseDir string) error
BuildChromeArgs(sessionURL, extensionPath string) []string
// alias BuildChromeLaunchArgs also acceptable — harness tries BuildChromeArgs first
```

- Layout `{BaseDir}/extension/{version}/manifest.json`
- Chrome args: include `--load-extension=<path>`; **omit** `--user-data-dir`
- Install stdout: absolute path, Developer mode, Load unpacked, `chrome://extensions`; ends `\n`
- Mini MV3 fixture embed OK for CI (`testdata/mini-extension/` shape)

**SPA HTML** (serve after extract/embed): root mount
`id="root"` or `data-browser-agent-root`; poll `/v1/session`; product id
`browser-agent`; port `43761`; install markers when not connected.

**React layout** under module root `react/` (preferred) — also accept
`project-api-capture-react/` only if `react/` missing? Prefer **`react/`** as
requirement states; implementer should create `react/`.

**Chrome-Ext-Browser-Agent/** at module root: `public/manifest.json` or
`manifest.json` with name referencing Browser Agent and **43761** in
host_permissions / content_scripts / externally_connectable.

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
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode values — top-level surface under test.
const (
	ModeCLIDispatch      = "cli-dispatch"
	ModeCLISidecmd       = "cli-sidecmd"
	ModeProductConfig    = "product-config"
	ModeSPAEmbed         = "spa-embed"
	ModeExtensionExtract = "extension-extract"
	ModeChromeArgs       = "chrome-args"
	ModeReactSrc         = "react-src"
	ModeExtShell         = "ext-shell"
)

// ExtractOp for ModeExtensionExtract.
const (
	ExtractOpFirst       = "first"
	ExtractOpRe          = "re"
	ExtractOpManifest    = "manifest-hosts"
	ExtractOpInstallCLI  = "install-cli"
)

// ReactProbe for ModeReactSrc.
const (
	ReactProbeProducts          = "products-browser-agent"
	ReactProbeApps              = "apps-entries"
	ReactProbeInstallGuideline  = "install-guideline"
)

// Sidecmd for ModeCLISidecmd.
const (
	SidecmdEval = "eval"
	SidecmdInfo = "info"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// ModuleRoot is project-api-capture module directory (filesystem leaves).
	// Root Setup sets from DOCTEST_ROOT/../..
	ModuleRoot string

	// BaseDir is temp parent for extract / serve state.
	BaseDir string

	// Session / server (sidecmd + spa-embed)
	Addr         string
	SessionID    string
	NoOpenChrome bool
	NoAgentRun   bool
	ReadyTimeout time.Duration

	// --- CLI (dispatch + sidecmd) ---
	CLIArgs         []string
	CLIEnv          map[string]string // nil keys treated as unset when map non-nil
	MaxDispatchWait time.Duration
	// Sidecmd selects eval vs info under ModeCLISidecmd.
	Sidecmd string
	// EvalExpr is the expression for eval (default "1+1").
	EvalExpr string
	// FakeExtension enables auto-complete WS for sidecmd eval.
	FakeExtension bool

	// --- product-config ---
	// ProductID: "browser-agent" | "browser-trace"
	ProductID string

	// --- spa-embed ---
	// SPAProbe: "root-hooks" | "install-guideline"
	SPAProbe string

	// --- extension-extract ---
	ExtractOp     string
	ExtractPasses int // 1 default; 2 for re-extract

	// --- chrome-args ---
	SessionURL    string
	ExtensionPath string // empty → extract first

	// --- react-src ---
	ReactProbe string
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// CLI shared
	Stdout   string
	Stderr   string
	ExitCode int
	ErrText  string
	// CLIErr is the error returned by HandleCLI (may be non-nil for bare usage).
	CLIErr string
	// DispatchTimedOut true if HandleCLI exceeded MaxDispatchWait.
	DispatchTimedOut bool

	// ProductConfig snapshot
	ProductID          string
	ProductDisplayName string
	ProductCLIName     string
	ProductControlPort int
	ProductPortStr     string
	ProductFeatures    []string
	ProductPageMarker  string
	ProductExtDirName  string

	// Extract
	InstallPath           string
	Version               string
	SecondPassInstallPath string
	SecondPassVersion     string
	ManifestPath          string
	ManifestText          string

	// Chrome args
	ChromeArgs []string

	// HTTP SPA
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string
	BaseURL     string
	ProbeURL    string
	RealSessionID string

	// Filesystem probes (react-src / ext-shell)
	FoundPaths    []string
	FileContents  map[string]string
	FileExists    bool
	CombinedText  string
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
	case ModeCLIDispatch:
		return runCLIDispatch(t, req)
	case ModeCLISidecmd:
		return runCLISidecmd(t, req)
	case ModeProductConfig:
		return runProductConfig(t, req)
	case ModeSPAEmbed:
		return runSPAEmbed(t, req)
	case ModeExtensionExtract:
		return runExtensionExtract(t, req)
	case ModeChromeArgs:
		return runChromeArgs(t, req)
	case ModeReactSrc:
		return runReactSrc(t, req)
	case ModeExtShell:
		return runExtShell(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- CLI dispatch (no server) ---

func runCLIDispatch(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	return invokeHandleCLI(t, req)
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
		// Explicit empty map so session resolve does not pick up ambient process env.
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
			// Return nil transport error so Assert can inspect CLIErr for expected failures.
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

// --- CLI side-command against live server ---

func runCLISidecmd(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Sidecmd == "" {
		t.Fatal("Sidecmd must be set (eval|info)")
	}
	srv, cleanup, err := startAgentServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// Fake extension for eval completion.
	var ext *fakeExtension
	if req.FakeExtension || req.Sidecmd == SidecmdEval {
		ext, err = dialFakeExtension(srv.BaseURL, "1.0.0", []string{"browser-agent"})
		if err != nil {
			return &Response{BaseURL: srv.BaseURL, RealSessionID: srv.SessionID},
				fmt.Errorf("fake extension dial: %w", err)
		}
		defer ext.Close()
		ext.AutoCompleteOK = true
		// Provide a simple data payload for eval.
		ext.ResultData = map[string]any{"value": 2, "result": 2, "ok": true}
		go ext.Loop()
		time.Sleep(40 * time.Millisecond)
	}

	expr := req.EvalExpr
	if expr == "" {
		expr = "1+1"
	}
	// Build CLI args: prefer leaf CLIArgs if fully set; else construct nested session <cmd>.
	args := req.CLIArgs
	if len(args) == 0 {
		switch req.Sidecmd {
		case SidecmdEval:
			args = []string{
				"session", "eval",
				"--session-id", srv.SessionID,
				"--addr", srv.BaseURL,
				expr,
			}
		case SidecmdInfo:
			args = []string{
				"session", "info",
				"--session-id", srv.SessionID,
				"--addr", srv.BaseURL,
				"--json",
			}
		default:
			return nil, fmt.Errorf("unknown Sidecmd %q", req.Sidecmd)
		}
	} else {
		// Ensure --addr / session present when leaf only set partial args.
		args = injectAddrAndSession(args, srv.BaseURL, srv.SessionID)
	}

	req2 := *req
	req2.CLIArgs = args
	if req2.CLIEnv == nil {
		req2.CLIEnv = map[string]string{}
	}
	if req2.MaxDispatchWait <= 0 {
		req2.MaxDispatchWait = 8 * time.Second
	}
	resp, err := invokeHandleCLI(t, &req2)
	if resp != nil {
		resp.BaseURL = srv.BaseURL
		resp.RealSessionID = srv.SessionID
	}
	return resp, err
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
	// Insert after "session <cmd>" when nested; else after first token.
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

// --- ProductConfig ---

func runProductConfig(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	id := req.ProductID
	if id == "" {
		t.Fatal("ProductID must be set")
	}
	resp := &Response{}
	switch id {
	case "browser-agent":
		pc := browseragent.ProductBrowserAgent
		fillProductResp(resp, pc)
	case "browser-trace":
		pc := browseragent.ProductBrowserTrace
		fillProductResp(resp, pc)
	default:
		return nil, fmt.Errorf("unknown ProductID %q", id)
	}
	return resp, nil
}

func fillProductResp(resp *Response, pc browseragent.ProductConfig) {
	resp.ProductID = pc.ID
	resp.ProductDisplayName = pc.DisplayName
	resp.ProductCLIName = pc.CLIName
	resp.ProductControlPort = pc.ControlPort
	if pc.ControlPort != 0 {
		resp.ProductPortStr = fmt.Sprintf("%d", pc.ControlPort)
	}
	resp.ProductFeatures = append([]string{}, pc.Features...)
	resp.ProductPageMarker = pc.PageMarkerGlobal
	resp.ProductExtDirName = pc.ExtensionDirName
}

// --- SPA embed ---

func runSPAEmbed(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startAgentServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		RealSessionID: srv.SessionID,
		BaseURL:       srv.BaseURL,
	}
	// Prefer /go; fall back to /.
	u := srv.BaseURL + "/go"
	if srv.SessionID != "" {
		u = u + "?session=" + url.QueryEscape(srv.SessionID)
	}
	status, ct, body, err := doGET(u)
	if err != nil {
		return resp, err
	}
	if status == http.StatusNotFound {
		u2 := srv.BaseURL + "/"
		status, ct, body, err = doGET(u2)
		if err != nil {
			return resp, err
		}
		u = u2
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = body
	resp.BodyString = string(body)
	resp.ProbeURL = u
	return resp, nil
}

// --- Extension extract ---

func runExtensionExtract(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set")
	}
	op := req.ExtractOp
	if op == "" {
		op = ExtractOpFirst
	}

	switch op {
	case ExtractOpInstallCLI:
		var stdout, stderr bytes.Buffer
		err := browseragent.InstallChromeExtension(&stdout, req.BaseDir)
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
		// InstallChromeExtension uses the canonical managed-chrome layout under
		// $HOME (baseDir is ignored); InstallPath must match that path for asserts.
		if p, v, e := browseragent.EnsureCanonicalExtension(); e == nil {
			resp.InstallPath = p
			resp.Version = v
			resp.ManifestPath = filepath.Join(p, "manifest.json")
		}
		return resp, nil

	case ExtractOpFirst, ExtractOpRe, ExtractOpManifest:
		passes := req.ExtractPasses
		if passes <= 0 {
			if op == ExtractOpRe {
				passes = 2
			} else {
				passes = 1
			}
		}
		path1, ver1, err := browseragent.ExtractEmbeddedExtension(req.BaseDir)
		if err != nil {
			return &Response{ExitCode: 1, ErrText: err.Error()}, err
		}
		resp := &Response{
			InstallPath:  path1,
			Version:      ver1,
			ManifestPath: filepath.Join(path1, "manifest.json"),
			ExitCode:     0,
		}
		if data, rerr := os.ReadFile(resp.ManifestPath); rerr == nil {
			resp.ManifestText = string(data)
		}
		if passes >= 2 {
			path2, ver2, err2 := browseragent.ExtractEmbeddedExtension(req.BaseDir)
			if err2 != nil {
				resp.ExitCode = 1
				resp.ErrText = err2.Error()
				return resp, err2
			}
			resp.SecondPassInstallPath = path2
			resp.SecondPassVersion = ver2
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown ExtractOp %q", op)
	}
}

// --- Chrome args ---

func runChromeArgs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	extPath := req.ExtensionPath
	if extPath == "" {
		p, _, err := browseragent.ExtractEmbeddedExtension(req.BaseDir)
		if err != nil {
			return &Response{ExitCode: 1, ErrText: err.Error()}, err
		}
		extPath = p
	}
	sessionURL := req.SessionURL
	if sessionURL == "" {
		sessionURL = "http://127.0.0.1:43761/go?session=test-sess"
	}
	args := browseragent.BuildChromeArgs(sessionURL, extPath)
	return &Response{
		InstallPath: extPath,
		ChromeArgs:  args,
		ExitCode:    0,
	}, nil
}

// --- React source filesystem ---

func runReactSrc(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	root := req.ModuleRoot
	if root == "" {
		t.Fatal("ModuleRoot empty")
	}
	resp := &Response{FileContents: map[string]string{}}
	reactRoot := filepath.Join(root, "react")
	// Also probe legacy monorepo folder name if react/ missing (document only;
	// asserts prefer react/).
	if st, err := os.Stat(reactRoot); err != nil || !st.IsDir() {
		alt := filepath.Join(root, "project-api-capture-react")
		if st2, err2 := os.Stat(alt); err2 == nil && st2.IsDir() {
			reactRoot = alt
		}
	}

	switch req.ReactProbe {
	case ReactProbeProducts:
		// Prefer react/src/products/browser-agent.ts(.tsx|.js)
		candidates := []string{
			filepath.Join(reactRoot, "src", "products", "browser-agent.ts"),
			filepath.Join(reactRoot, "src", "products", "browser-agent.tsx"),
			filepath.Join(reactRoot, "src", "products", "browser-agent.js"),
		}
		path, data, ok := firstExistingFile(candidates)
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		}
		return resp, nil

	case ReactProbeApps:
		// session-page + popup entry points
		sessionCandidates := []string{
			filepath.Join(reactRoot, "src", "apps", "session-page", "main.tsx"),
			filepath.Join(reactRoot, "src", "apps", "session-page", "main.ts"),
			filepath.Join(reactRoot, "src", "apps", "session-page", "main.jsx"),
			filepath.Join(reactRoot, "src", "apps", "session-page", "index.tsx"),
		}
		popupCandidates := []string{
			filepath.Join(reactRoot, "src", "apps", "popup", "main.tsx"),
			filepath.Join(reactRoot, "src", "apps", "popup", "main.ts"),
			filepath.Join(reactRoot, "src", "apps", "popup", "main.jsx"),
			filepath.Join(reactRoot, "src", "apps", "popup", "index.tsx"),
		}
		sp, _, okS := firstExistingFile(sessionCandidates)
		pp, _, okP := firstExistingFile(popupCandidates)
		resp.FileExists = okS && okP
		if okS {
			resp.FoundPaths = append(resp.FoundPaths, sp)
		}
		if okP {
			resp.FoundPaths = append(resp.FoundPaths, pp)
		}
		return resp, nil

	case ReactProbeInstallGuideline:
		candidates := []string{
			filepath.Join(reactRoot, "src", "ui", "InstallGuideline.tsx"),
			filepath.Join(reactRoot, "src", "ui", "InstallGuideline.ts"),
			filepath.Join(reactRoot, "src", "ui", "InstallGuideline.jsx"),
			filepath.Join(reactRoot, "src", "components", "InstallGuideline.tsx"),
		}
		path, data, ok := firstExistingFile(candidates)
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown ReactProbe %q", req.ReactProbe)
	}
}

// --- Extension shell on disk ---

func runExtShell(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	root := req.ModuleRoot
	extDir := filepath.Join(root, "Chrome-Ext-Browser-Agent")
	resp := &Response{FileContents: map[string]string{}}

	if st, err := os.Stat(extDir); err != nil || !st.IsDir() {
		resp.FileExists = false
		resp.ErrText = fmt.Sprintf("Chrome-Ext-Browser-Agent dir missing under %s", root)
		return resp, nil
	}
	resp.FoundPaths = append(resp.FoundPaths, extDir)

	manifestCandidates := []string{
		filepath.Join(extDir, "public", "manifest.json"),
		filepath.Join(extDir, "manifest.json"),
		filepath.Join(extDir, "src", "manifest.json"),
		filepath.Join(extDir, "build", "manifest.json"),
	}
	path, data, ok := firstExistingFile(manifestCandidates)
	resp.FileExists = ok
	if ok {
		resp.FoundPaths = append(resp.FoundPaths, path)
		resp.ManifestPath = path
		resp.ManifestText = string(data)
		resp.FileContents[path] = string(data)
		resp.CombinedText = string(data)
	}
	return resp, nil
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

// --- server harness (shared with browser-agent style) ---

type agentServer struct {
	BaseURL   string
	SessionID string
	cancel    context.CancelFunc
	done      <-chan error
}

func startAgentServer(t *testing.T, req *Request) (*agentServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	if req.SessionID == "" {
		t.Fatal("SessionID must be set by Setup")
	}
	addr := req.Addr
	if addr == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, nil, err
		}
		addr = ln.Addr().String()
		_ = ln.Close()
		req.Addr = addr
	}
	ready := req.ReadyTimeout
	if ready <= 0 {
		ready = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	var stdout, stderr bytes.Buffer
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      req.BaseDir,
		SessionID:    req.SessionID,
		NoOpenChrome: true,
		NoAgentRun:   true,
		Stdout:       &stdout,
		Stderr:       &stderr,
	}

	done := make(chan error, 1)
	go func() {
		_, err := browseragent.Run(ctx, cfg)
		done <- err
	}()

	baseURL := "http://" + addr
	if err := waitHealth(baseURL, ready); err != nil {
		cancel()
		<-done
		return nil, nil, fmt.Errorf("control server never healthy at %s: %w", baseURL, err)
	}

	srv := &agentServer{
		BaseURL:   baseURL,
		SessionID: req.SessionID,
		cancel:    cancel,
		done:      done,
	}
	cleanup := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
	}
	return srv, cleanup, nil
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

func doGET(rawURL string) (int, string, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, "", nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, res.Header.Get("Content-Type"), nil, err
	}
	return res.StatusCode, res.Header.Get("Content-Type"), body, nil
}

// --- fake extension WS client ---

type wsEnvelope struct {
	V       int            `json:"v"`
	Type    string         `json:"type"`
	ID      string         `json:"id,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

type fakeExtension struct {
	conn           *websocket.Conn
	AutoCompleteOK bool
	ResultData     map[string]any
	OnJob          func(wsEnvelope)
	JobsSeen       int
	version        string
	features       []string
}

func dialFakeExtension(baseURL, version string, features []string) (*fakeExtension, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
	dialer := websocket.Dialer{HandshakeTimeout: 2 * time.Second}
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
	return &fakeExtension{
		conn:     conn,
		version:  version,
		features: features,
	}, nil
}

func (f *fakeExtension) Close() {
	if f != nil && f.conn != nil {
		_ = f.conn.Close()
	}
}

func (f *fakeExtension) SendHello() error {
	env := wsEnvelope{
		V:    1,
		Type: "hello",
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
		var env wsEnvelope
		if err := f.conn.ReadJSON(&env); err != nil {
			return
		}
		if env.Type == "job" {
			f.JobsSeen++
			if f.OnJob != nil {
				f.OnJob(env)
			}
			if f.AutoCompleteOK {
				jobID := env.ID
				if jobID == "" && env.Payload != nil {
					if id, ok := env.Payload["id"].(string); ok {
						jobID = id
					} else if id, ok := env.Payload["job_id"].(string); ok {
						jobID = id
					}
				}
				data := f.ResultData
				if data == nil {
					data = map[string]any{"value": 2}
				}
				_ = f.conn.WriteJSON(wsEnvelope{
					V:    1,
					Type: "result",
					ID:   jobID,
					Payload: map[string]any{
						"job_id": jobID,
						"ok":     true,
						"data":   data,
					},
				})
			}
		}
	}
}
```
