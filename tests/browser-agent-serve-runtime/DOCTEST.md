# browser-agent serve runtime — Chrome + agent-run hooks + meta + extension WS

Exercises the **serve launcher / runtime** layer on package
`github.com/xhd2015/browser-agent/browseragent` (next slice after
sealed `tests/browser-agent` + `tests/browser-agent-cli-react`):

| Surface | What is under test |
|---------|-------------------|
| Serve artifacts | On serve: write `SYSTEM.md` + `meta.json` under session dir |
| Chrome open hook | `NoOpenChrome` / injectable `OpenChromeFn` (never real Chrome) |
| Agent-run hook | `NoAgentRun` / injectable `AgentRunFn` (never real agent-run) |
| BuildAgentRunArgs | Pure argv builder: prefixed `--session-id`, `--env BROWSER_AGENT_SESSION_ID`, `--no-submit`, … |
| Extension WS source | Shell + embedded mini sources contain WS hello/job/result tokens |
| Session install path | Optional: `GET /v1/session` exposes `extension_install_path` after extract |

**No real Chrome.** **No real agent-run.** Hooks record calls only when set.
Extension leaves are **filesystem/source** asserts (not live CDP).

**Sealed** (do not modify; must stay GREEN after implement):

```sh
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-cli-react/...
```

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** starts **`browser-agent serve`**. The **Serve Runtime** (package
`Run(ctx, Config)`) is a standby co-pilot launcher:

1. Create session dir `{BaseDir}/sessions/{sessionID}/`
2. Extract embedded Chrome extension under `{BaseDir}/extension/{version}/`
3. Write **SYSTEM.md** playbook (nested `browser-agent session …` recipes;
   **no** concrete control session id; mentions `BROWSER_AGENT_SESSION_ID`)
4. Write **meta.json** (session_id, addr/base_url/session_url, product,
   system_prompt_path, extension_install_path, control_port)
5. Listen on control **Addr** (default product `127.0.0.1:43761`)
6. Unless `NoOpenChrome`: open Chrome with session URL + load-extension path
7. Unless `NoAgentRun`: launch **agent-run** via `BuildAgentRunArgs(control, …)`
   (`--session-id=<agent-run-id>`, `--env BROWSER_AGENT_SESSION_ID=<control>`,
   `--no-submit`); **no** manual process env overlay for session id.
   Serve operator logs still print the **control** session id.

**Config hooks** (testability; production path when nil):

```text
Config {
  Addr, BaseDir, SessionID, WorkspaceDir?
  NoOpenChrome, NoAgentRun
  OpenChromeFn(sessionURL, extensionInstallPath) error   // nil → production
  AgentRunFn(sessionID, systemPromptPath, workspaceDir, env) error
}
```

- When `NoOpenChrome` / `NoAgentRun` is true → **must not** call the matching
  injector / production launcher.
- When flag is false and injector is set → call **once** after listen is ready
  with correct args; do **not** exec real binaries.
- When injector is nil and flag false → production best-effort launch; failures
  log warning and do **not** crash serve (not asserted in this tree).

**BuildAgentRunArgs** (pure; control id in, agent-run id + env flag out):

```text
BuildAgentRunArgs(controlSessionID, promptOrSystemPath, workspaceDir) []string
# must include (order flexible):
#   run
#   --session-id=<AgentRunSessionID(control)>   # browser-agent-sess-<control>
#   --agent-runner=grok-tty
#   --auto-send-or-resume
#   --new-terminal
#   optional --dir=<workspace> when workspace non-empty
#   --env BROWSER_AGENT_SESSION_ID=<control>    # control id for nested CLI resolve
#   --no-submit          # ALWAYS — first prompt stays draft (no auto-submit in TTY)
#   --open
#   --  <prompt referencing SYSTEM; must not require bare control id>
#
# Canonical shape (token order flexible; --no-submit always present):
#   agent-run run
#     --session-id=browser-agent-sess-<control>
#     --agent-runner=grok-tty
#     --auto-send-or-resume
#     --new-terminal
#     [--dir=<ws>]
#     --env BROWSER_AGENT_SESSION_ID=<control>
#     --no-submit
#     --open
#     --
#     <prompt>
```

**meta.json** (fields used by tests; extras OK):

```json
{
  "session_id": "...",
  "addr": "127.0.0.1:PORT",
  "base_url": "http://127.0.0.1:PORT",
  "session_url": "http://127.0.0.1:PORT/go?session=...",
  "system_prompt_path": ".../SYSTEM.md",
  "extension_install_path": ".../extension/1.0.0",
  "product": "browser-agent",
  "control_port": 43761
}
```

**Extension Agent source** (Chrome-Ext-Browser-Agent + embedded mini):

- Background connects `ws://127.0.0.1:43761/v1/ws` (or CONTROL_PORT / `/v1/ws`)
- Sends hello `{v:1, type:"hello", payload:{version, features:["browser-agent",…]}}`
- On job for `eval`/`info`: responds `type:"result"` with job id + ok/error/data
- Reconnect attempt present in source (retry / alarm / onclose)
- contentScript sets `window.__BROWSER_AGENT_EXT__` with features including
  `browser-agent`

**Test Client** in this tree:

- Serve leaves start `browseragent.Run` with temp BaseDir, free loopback Addr,
  known SessionID, injectable hooks, short settle after health, then cancel.
- Pure leaves call `BuildAgentRunArgs` only.
- Ext-source leaves read ModuleRoot filesystem (shell + embedded).
- Session-install-path leaf serves with hooks off and GETs `/v1/session`.

## Decision Tree

```
browser-agent-serve-runtime
├── serve-artifacts/                           [serve: SYSTEM.md + meta.json]
│   ├── system-md/                               A1 nested recipes; no control id on disk
│   └── meta-json/                               A2 session_id, urls, product
├── chrome-hook/                               [OpenChromeFn / NoOpenChrome]
│   ├── skipped/                                 B1 flag true → fn never called
│   └── called/                                  B2 flag false → once; URL + path
├── agent-hook/                                [AgentRunFn / NoAgentRun]
│   ├── skipped/                                 C1 flag true → fn never called
│   └── called/                                  C2 once; SYSTEM.md path; env overlay optional
├── agent-args/                                [pure BuildAgentRunArgs]
│   ├── with-workspace/                          D1 prefix session-id, --env, dir, --no-submit…
│   └── empty-workspace/                         D2 no --dir; still prefix + --env + --no-submit
├── ext-source/                                [filesystem WS protocol tokens]
│   ├── shell-background/                        E1 Chrome-Ext-Browser-Agent bg
│   ├── content-script/                          E2 __BROWSER_AGENT_EXT__ + feature
│   └── embedded-background/                     E3 mini embed WS agent tokens
└── session-install-path/                      [GET /v1/session optional F1]
    └── v1-session-has-path/                     F1 extension_install_path present
```

### Parameter significance (high → low)

1. **Surface / Mode** — serve artifacts vs chrome hook vs agent hook vs pure
   argv vs FS source vs session JSON (different `Run` contracts).
2. **Within hooks** — skipped vs called (largest behavioral split for launch).
3. **Within agent-args** — workspace non-empty vs empty (`--dir` presence).
4. **Within ext-source** — which file surface (shell bg / content / embed).
5. **Leaf details** — field strings, path suffixes, settle timing.

## Test Index

| Leaf | Scenario |
|------|----------|
| `serve-artifacts/system-md` | (A1) serve NoOpenChrome+NoAgentRun → SYSTEM.md nested recipes; no control id; env mention |
| `serve-artifacts/meta-json` | (A2) same serve → meta.json with session_id, base_url/session_url, product browser-agent |
| `chrome-hook/skipped` | (B1) NoOpenChrome=true + OpenChromeFn set → fn **not** called |
| `chrome-hook/called` | (B2) NoOpenChrome=false + OpenChromeFn records → called once; sessionURL has `/go` + session id; ext path non-empty |
| `agent-hook/skipped` | (C1) NoAgentRun=true + AgentRunFn set → fn **not** called |
| `agent-hook/called` | (C2) NoAgentRun=false + AgentRunFn records → once; control sessionID; system path ends with SYSTEM.md; env overlay optional |
| `agent-args/with-workspace` | (D1) argv: prefixed `--session-id`, `--env BROWSER_AGENT_SESSION_ID`, dir, **`--no-submit`**, open |
| `agent-args/empty-workspace` | (D2) empty workspace → no `--dir`; still prefix + `--env` + **`--no-submit`** |
| `ext-source/shell-background` | (E1) Chrome-Ext-Browser-Agent background has `/v1/ws` or `ws://` + hello + job/result |
| `ext-source/content-script` | (E2) contentScript sets `__BROWSER_AGENT_EXT__` and `browser-agent` feature |
| `ext-source/embedded-background` | (E3) embedded mini background (or extract) has WS agent protocol tokens |
| `session-install-path/v1-session-has-path` | (F1) after serve extract, GET `/v1/session` includes non-empty `extension_install_path` |

**Leaf count: 12**

## How to Run

```sh
cd project-api-capture
doctest vet ./tests/browser-agent-serve-runtime
doctest test ./tests/browser-agent-serve-runtime/...
# or:
cd tests/browser-agent-serve-runtime && doctest vet . && doctest test -v .
```

### Implementer contract (authoritative for GREEN)

```text
type Config struct {
    Addr         string
    BaseDir      string
    SessionID    string
    WorkspaceDir string // optional; passed to AgentRunFn / BuildAgentRunArgs
    NoOpenChrome bool
    NoAgentRun   bool
    Stdout       io.Writer
    Stderr       io.Writer
    ReadyTimeout time.Duration

    // Optional injectors (nil → production best-effort; never used when flags true)
    OpenChromeFn func(sessionURL, extensionInstallPath string) error
    AgentRunFn   func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error
}

func AgentRunSessionID(controlID string) string // browser-agent-sess- + control (idempotent)

func BuildAgentRunArgs(controlSessionID, promptOrSystemPath, workspaceDir string) []string
// ALWAYS includes --no-submit so serve-launched agent-run opens with a draft
// first prompt (no auto-submit in TTY). Prefer bare token "--no-submit".
// --session-id=<AgentRunSessionID(control)>
// --env BROWSER_AGENT_SESSION_ID=<control>
// Optional --dir only when workspaceDir non-empty. Production launchAgentRun
// must use this argv (no separate path that omits --no-submit); no manual
// cmd.Env overlay for session id.
```

Serve sequence after listen succeeds:

1. extract extension (if not already)
2. write SYSTEM.md + meta.json (SYSTEM.md without concrete control id)
3. if !NoOpenChrome: call OpenChromeFn or production open once
4. if !NoAgentRun: call AgentRunFn or production launch once with control
   sessionID as BuildAgentRunArgs input (session carried via argv `--env`)

Session JSON may include `extension_install_path` (snake_case preferred;
harness also accepts camelCase). meta.json path:
`{BaseDir}/sessions/{sessionID}/meta.json`.

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

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode values — top-level surface under test.
const (
	ModeServeArtifacts      = "serve-artifacts"
	ModeChromeHook          = "chrome-hook"
	ModeAgentHook           = "agent-hook"
	ModeAgentArgs           = "agent-args"
	ModeExtSource           = "ext-source"
	ModeSessionInstallPath  = "session-install-path"
)

// ServeArtifactProbe for ModeServeArtifacts.
const (
	ServeProbeSystemMD = "system-md"
	ServeProbeMetaJSON = "meta-json"
)

// HookExpect for chrome-hook / agent-hook.
const (
	HookExpectSkipped = "skipped"
	HookExpectCalled  = "called"
)

// ExtSourceTarget for ModeExtSource.
const (
	ExtSrcShellBackground     = "shell-background"
	ExtSrcContentScript       = "content-script"
	ExtSrcEmbeddedBackground  = "embedded-background"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// ModuleRoot is project-api-capture module directory (filesystem leaves).
	// Root Setup sets from DOCTEST_ROOT/../..
	ModuleRoot string

	// BaseDir is temp parent for extract / session / serve state.
	BaseDir string

	// Server / session
	Addr         string
	SessionID    string
	WorkspaceDir string
	NoOpenChrome bool
	NoAgentRun   bool
	ReadyTimeout time.Duration
	// HookSettle is how long to wait after health before reading hook records.
	HookSettle time.Duration

	// InjectOpenChromeFn / InjectAgentRunFn: harness installs recording hooks.
	InjectOpenChromeFn bool
	InjectAgentRunFn   bool

	// ServeArtifactProbe: system-md | meta-json
	ServeArtifactProbe string

	// HookExpect: skipped | called
	HookExpect string

	// --- agent-args (pure) ---
	AgentArgsSessionID  string
	AgentArgsPromptPath string
	AgentArgsWorkspace  string

	// --- ext-source ---
	ExtSourceTarget string
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// Serve / session paths
	SessionDir         string
	SystemMDPath       string
	SystemMDText       string
	MetaPath           string
	MetaJSON           string
	Meta               map[string]any
	RealSessionID      string
	BaseURL            string
	Addr               string
	ExtensionInstallPath string

	// Hook recordings
	OpenChromeCallCount int
	OpenChromeSessionURL string
	OpenChromeExtPath    string
	AgentRunCallCount    int
	AgentRunSessionID    string
	AgentRunSystemPath   string
	AgentRunWorkspace    string
	AgentRunEnv          map[string]string

	// BuildAgentRunArgs
	AgentRunArgs []string

	// HTTP session probe
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string
	ProbeURL    string
	Raw         map[string]any
	// SessionJSONExtensionInstallPath from GET /v1/session
	SessionJSONExtensionInstallPath string
	SessionIDField                  string

	// Filesystem probes (ext-source)
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	ErrText      string
	ExitCode     int
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
	case ModeServeArtifacts:
		return runServeArtifacts(t, req)
	case ModeChromeHook:
		return runChromeHook(t, req)
	case ModeAgentHook:
		return runAgentHook(t, req)
	case ModeAgentArgs:
		return runAgentArgs(t, req)
	case ModeExtSource:
		return runExtSource(t, req)
	case ModeSessionInstallPath:
		return runSessionInstallPath(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- serve artifacts ---

func runServeArtifacts(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	// Always skip real launch for artifact leaves.
	req.NoOpenChrome = true
	req.NoAgentRun = true
	req.InjectOpenChromeFn = false
	req.InjectAgentRunFn = false

	srv, cleanup, rec, err := startServe(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := baseServeResp(srv, rec)
	if err := fillSessionArtifacts(resp, req); err != nil {
		return resp, err
	}
	return resp, nil
}

// --- chrome hook ---

func runChromeHook(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.HookExpect == "" {
		t.Fatal("HookExpect must be set (skipped|called)")
	}
	// Isolate chrome path: never launch agent-run in this mode.
	req.NoAgentRun = true
	req.InjectAgentRunFn = false
	req.InjectOpenChromeFn = true
	switch req.HookExpect {
	case HookExpectSkipped:
		req.NoOpenChrome = true
	case HookExpectCalled:
		req.NoOpenChrome = false
	default:
		return nil, fmt.Errorf("unknown HookExpect %q", req.HookExpect)
	}

	srv, cleanup, rec, err := startServe(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := baseServeResp(srv, rec)
	_ = fillSessionArtifacts(resp, req)
	return resp, nil
}

// --- agent hook ---

func runAgentHook(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.HookExpect == "" {
		t.Fatal("HookExpect must be set (skipped|called)")
	}
	// Isolate agent-run path: never open chrome in this mode.
	req.NoOpenChrome = true
	req.InjectOpenChromeFn = false
	req.InjectAgentRunFn = true
	switch req.HookExpect {
	case HookExpectSkipped:
		req.NoAgentRun = true
	case HookExpectCalled:
		req.NoAgentRun = false
	default:
		return nil, fmt.Errorf("unknown HookExpect %q", req.HookExpect)
	}
	if req.WorkspaceDir == "" {
		req.WorkspaceDir = filepath.Join(req.BaseDir, "workspace")
		_ = os.MkdirAll(req.WorkspaceDir, 0o755)
	}

	srv, cleanup, rec, err := startServe(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := baseServeResp(srv, rec)
	_ = fillSessionArtifacts(resp, req)
	return resp, nil
}

// --- pure BuildAgentRunArgs ---

func runAgentArgs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	sid := req.AgentArgsSessionID
	if sid == "" {
		sid = "sess-agent-args"
	}
	prompt := req.AgentArgsPromptPath
	if prompt == "" {
		prompt = "/tmp/sessions/" + sid + "/SYSTEM.md"
	}
	args := browseragent.BuildAgentRunArgs(sid, prompt, req.AgentArgsWorkspace)
	return &Response{
		RealSessionID: sid,
		AgentRunArgs:  args,
		ExitCode:      0,
	}, nil
}

// --- extension source filesystem ---

func runExtSource(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ExtSourceTarget == "" {
		t.Fatal("ExtSourceTarget must be set")
	}
	root := req.ModuleRoot
	resp := &Response{FileContents: map[string]string{}}

	switch req.ExtSourceTarget {
	case ExtSrcShellBackground:
		// Prefer public/background.js; also accept root/src/build names.
		candidates := []string{
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "background.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "background.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "background.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "background.js"),
		}
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

	case ExtSrcContentScript:
		candidates := []string{
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "contentScript.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "content.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "contentScript.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "contentScript.js"),
			filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "contentScript.js"),
		}
		path, data, ok := firstExistingFile(candidates)
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "contentScript not found under Chrome-Ext-Browser-Agent"
		}
		return resp, nil

	case ExtSrcEmbeddedBackground:
		// Prefer package embedded sources; fall back to extract under BaseDir.
		candidates := []string{
			filepath.Join(root, "browseragent", "embedded", "extension", "background.js"),
			filepath.Join(root, "browseragent", "embedded", "extension", "background", "service_worker.js"),
			filepath.Join(root, "browseragent", "embedded", "extension", "sw.js"),
		}
		path, data, ok := firstExistingFile(candidates)
		if !ok && req.BaseDir != "" {
			// Extract then read.
			if installPath, _, err := browseragent.ExtractEmbeddedExtension(req.BaseDir); err == nil {
				extractCandidates := []string{
					filepath.Join(installPath, "background.js"),
					filepath.Join(installPath, "service_worker.js"),
					filepath.Join(installPath, "sw.js"),
				}
				path, data, ok = firstExistingFile(extractCandidates)
				if ok {
					resp.ExtensionInstallPath = installPath
				}
			}
		}
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "embedded extension background not found (package embed or extract)"
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown ExtSourceTarget %q", req.ExtSourceTarget)
	}
}

// --- GET /v1/session extension_install_path ---

func runSessionInstallPath(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	req.NoOpenChrome = true
	req.NoAgentRun = true
	req.InjectOpenChromeFn = false
	req.InjectAgentRunFn = false

	srv, cleanup, rec, err := startServe(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := baseServeResp(srv, rec)
	_ = fillSessionArtifacts(resp, req)

	u := srv.BaseURL + "/v1/session?session=" + url.QueryEscape(srv.SessionID)
	status, ct, body, gerr := doGET(u)
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = body
	resp.BodyString = string(body)
	resp.ProbeURL = u
	if gerr != nil {
		return resp, gerr
	}
	parseSessionInstallPath(resp, body)
	return resp, nil
}

// --- serve harness ---

type hookRecorder struct {
	mu sync.Mutex

	openChromeN    int
	openChromeURL  string
	openChromePath string

	agentRunN      int
	agentRunSID    string
	agentRunSys    string
	agentRunWS     string
	agentRunEnv    map[string]string
}

type agentServer struct {
	BaseURL   string
	SessionID string
	Addr      string
	cancel    context.CancelFunc
	done      <-chan error
}

func startServe(t *testing.T, req *Request) (*agentServer, func(), *hookRecorder, error) {
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
			return nil, nil, nil, err
		}
		addr = ln.Addr().String()
		_ = ln.Close()
		req.Addr = addr
	}
	ready := req.ReadyTimeout
	if ready <= 0 {
		ready = 5 * time.Second
	}
	settle := req.HookSettle
	if settle <= 0 {
		settle = 120 * time.Millisecond
	}

	rec := &hookRecorder{agentRunEnv: map[string]string{}}
	ctx, cancel := context.WithCancel(context.Background())
	var stdout, stderr bytes.Buffer
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      req.BaseDir,
		SessionID:    req.SessionID,
		NoOpenChrome: req.NoOpenChrome,
		NoAgentRun:   req.NoAgentRun,
		Stdout:       &stdout,
		Stderr:       &stderr,
	}
	// WorkspaceDir if Config supports it (optional field — set via helper when present).
	setConfigWorkspaceDir(&cfg, req.WorkspaceDir)

	if req.InjectOpenChromeFn {
		cfg.OpenChromeFn = func(sessionURL, extensionInstallPath string) error {
			rec.mu.Lock()
			defer rec.mu.Unlock()
			rec.openChromeN++
			if rec.openChromeN == 1 {
				rec.openChromeURL = sessionURL
				rec.openChromePath = extensionInstallPath
			}
			return nil
		}
	}
	if req.InjectAgentRunFn {
		cfg.AgentRunFn = func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error {
			rec.mu.Lock()
			defer rec.mu.Unlock()
			rec.agentRunN++
			if rec.agentRunN == 1 {
				rec.agentRunSID = sessionID
				rec.agentRunSys = systemPromptPath
				rec.agentRunWS = workspaceDir
				rec.agentRunEnv = cloneStringMap(env)
			}
			return nil
		}
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
		return nil, nil, rec, fmt.Errorf("control server never healthy at %s: %w", baseURL, err)
	}
	// Allow serve-side post-listen hooks to fire.
	time.Sleep(settle)

	srv := &agentServer{
		BaseURL:   baseURL,
		SessionID: req.SessionID,
		Addr:      addr,
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
	return srv, cleanup, rec, nil
}

func baseServeResp(srv *agentServer, rec *hookRecorder) *Response {
	resp := &Response{
		RealSessionID: srv.SessionID,
		BaseURL:       srv.BaseURL,
		Addr:          srv.Addr,
		ExitCode:      0,
	}
	if rec != nil {
		rec.mu.Lock()
		resp.OpenChromeCallCount = rec.openChromeN
		resp.OpenChromeSessionURL = rec.openChromeURL
		resp.OpenChromeExtPath = rec.openChromePath
		resp.AgentRunCallCount = rec.agentRunN
		resp.AgentRunSessionID = rec.agentRunSID
		resp.AgentRunSystemPath = rec.agentRunSys
		resp.AgentRunWorkspace = rec.agentRunWS
		resp.AgentRunEnv = cloneStringMap(rec.agentRunEnv)
		rec.mu.Unlock()
	}
	return resp
}

func fillSessionArtifacts(resp *Response, req *Request) error {
	sid := req.SessionID
	if sid == "" {
		sid = resp.RealSessionID
	}
	sessionDir := filepath.Join(req.BaseDir, "sessions", sid)
	resp.SessionDir = sessionDir
	resp.SystemMDPath = filepath.Join(sessionDir, "SYSTEM.md")
	resp.MetaPath = filepath.Join(sessionDir, "meta.json")

	if data, err := os.ReadFile(resp.SystemMDPath); err == nil {
		resp.SystemMDText = string(data)
	}
	if data, err := os.ReadFile(resp.MetaPath); err == nil {
		resp.MetaJSON = string(data)
		var m map[string]any
		if json.Unmarshal(data, &m) == nil {
			resp.Meta = m
			if p := stringField(m, "extension_install_path", "extensionInstallPath"); p != "" {
				resp.ExtensionInstallPath = p
			}
		}
	}
	return nil
}

func parseSessionInstallPath(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.Raw = raw
	if id, ok := raw["session_id"].(string); ok {
		resp.SessionIDField = id
	}
	// Prefer top-level; also accept nested meta.
	if p := stringField(raw, "extension_install_path", "extensionInstallPath"); p != "" {
		resp.SessionJSONExtensionInstallPath = p
		return
	}
	if nested, ok := raw["meta"].(map[string]any); ok {
		if p := stringField(nested, "extension_install_path", "extensionInstallPath"); p != "" {
			resp.SessionJSONExtensionInstallPath = p
		}
	}
}

func stringField(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return t
				}
			}
		}
	}
	return ""
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

func firstExistingFile(paths []string) (string, []byte, bool) {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return p, data, true
		}
	}
	return "", nil, false
}

func cloneStringMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// setConfigWorkspaceDir sets WorkspaceDir when Config has that field.
// Compile-time: Config.WorkspaceDir is part of the implementer contract.
func setConfigWorkspaceDir(cfg *browseragent.Config, workspace string) {
	if strings.TrimSpace(workspace) == "" {
		return
	}
	cfg.WorkspaceDir = workspace
}

// argv helpers used by Assert packages (shared via root helpers in SETUP.md).
func argvJoined(args []string) string {
	return strings.Join(args, "\x00")
}

func argvHasToken(args []string, token string) bool {
	for _, a := range args {
		if a == token || strings.HasPrefix(a, token+"=") {
			return true
		}
	}
	return false
}

func argvHasFlagValue(args []string, flag, want string) bool {
	for i, a := range args {
		if a == flag && i+1 < len(args) && args[i+1] == want {
			return true
		}
		if strings.HasPrefix(a, flag+"=") && strings.TrimPrefix(a, flag+"=") == want {
			return true
		}
	}
	return false
}

func argvHasFlagWithAnyValue(args []string, flag string) bool {
	for i, a := range args {
		if a == flag && i+1 < len(args) && strings.TrimSpace(args[i+1]) != "" {
			return true
		}
		if strings.HasPrefix(a, flag+"=") && strings.TrimSpace(strings.TrimPrefix(a, flag+"=")) != "" {
			return true
		}
	}
	return false
}

func argvHasDirFlag(args []string) bool {
	return argvHasFlagWithAnyValue(args, "--dir")
}
```
