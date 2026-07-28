# browser-agent Firefox connect agent + bundle embed (Phase 1 A + E1)

Classic TDD for **Phase 1** of Firefox support: content-script **register**, background
**WebSocket + hello** (with `browser_product: "firefox"`), real **keepalive/reconnect**,
and bundle staging into `browseragent/embedded/extension-firefox`.

| Surface | What is under test |
|---------|-------------------|
| Content script | `Firefox-Ext-Browser-Agent/public/contentScript.js` register message |
| Background WS | `background.js` WebSocket `/v1/ws?session=`, hello, browser_product |
| Background keepalive | reconnect / keepalive / ping (not no-op-only 1min alarm) |
| Bundle E1 | Stage Firefox public→build→`embedded/extension-firefox` (temp Root) |
| Chrome regression | `BuildExtensionShell` still Chrome-only |

**No real Firefox. No live WebSocket. No network.** Source markers + pure Go bundle APIs only.

**Classic TDD — RED** until implementer ports Chrome agent protocol connect path into
Firefox-Ext public and wires bundle Firefox embed staging.

### Out of scope (later phases)

- Full job handlers (eval/cdp/screenshot) beyond connect
- SPA Firefox install panel (Phase 4)
- Chrome path changes (regression only)

## Version

0.0.2

# DSN (Domain Specific Notion)

**Session Page** (`GET /go?session=<id>`) is opened in Firefox. The **Content Script**
runs at `document_start`, keeps the `__BROWSER_AGENT_EXT__` marker, parses session from
`/go?session=` or `data-session-id`, and sends **`register`** to the background via
`browser.runtime.sendMessage` (or `chrome.runtime.sendMessage` polyfill).

**Background Worker** maintains a **sessions map**. On register it opens
`ws://host:port/v1/ws?session=<id>`, sends **hello** with version, features including
`browser-agent`, optional `bundle_md5`, and **`browser_product: "firefox"`**, handles
**ping→pong**, reconnects dead sockets (alarm + backoff), and runs **keepalive** while
WS is open (not a no-op-only 1-minute alarm stub).

**Bundle pipeline** stages Firefox shell into the embed tree used by `//go:embed`:

```text
Firefox-Ext-Browser-Agent/public
  -> BuildFirefoxExtensionShell (public→build)
  -> StageFirefoxExtensionEmbed / Bundle
  -> {Root}/browseragent/embedded/extension-firefox/
```

Fixture mode may stage `browseragent/fixtures/extension-firefox` into the same embed rel.

**Test Client** reads Firefox-Ext public sources and calls pure Go APIs; never launches
a browser.

```text
Content Script on /go?session=S
  -> __BROWSER_AGENT_EXT__ marker retained
  -> browser.runtime.sendMessage({type:"register", session_id:S, control_port, …})

Background on register
  -> sessions[S] = {ws, tabId, windowId, controlPort, …}
  -> new WebSocket(ws://host:port/v1/ws?session=S)
  -> hello { version, features:["browser-agent",…], browser_product:"firefox", bundle_md5? }
  -> keepalive ping while OPEN; alarm reconnects dead sockets

Bundle / StageFirefoxExtensionEmbed(Root)
  -> {Root}/browseragent/embedded/extension-firefox/manifest.json
```

## Decision Tree

```
browser-agent-firefox-connect
├── ext-source/                                  [Firefox-Ext-Browser-Agent/public markers]
│   ├── content-script-register/                   register + session_id + sendMessage; marker
│   ├── background-hello-ws/                       WebSocket, hello, /v1/ws, browser_product|firefox
│   └── background-keepalive/                      reconnect/keepalive/ping; not no-op-only alarm
├── bundle-stages-firefox-embed/                 [StageFirefox / Bundle → extension-firefox]
│   └── (leaf)                                     temp Root; staged embed has manifest
└── chrome-regression/                           [Chrome paths unchanged]
    └── build-extension-shell-chrome/              BuildExtensionShell still Chrome-Ext only
```

### Parameter significance (high → low)

1. **Surface / Mode** — ext-source vs bundle embed staging vs chrome regression
   (largest behavioral split; different Run branches).
2. **Within ext-source** — content-script register vs background hello/WS vs keepalive
   (distinct files / contracts).
3. **Within chrome-regression** — BuildExtensionShell chrome-only smoke.

## Test Index

| Leaf | Scenario |
|------|----------|
| `ext-source/content-script-register` | `public/contentScript.js` has register / session_id / sendMessage (browser or chrome.runtime); keeps `__BROWSER_AGENT_EXT__`; reads session from go page / data-session-id |
| `ext-source/background-hello-ws` | `public/background.js` has WebSocket, hello, `/v1/ws`, `browser_product` or firefox |
| `ext-source/background-keepalive` | reconnect / keepalive / ping language present; not only empty no-op alarm stub |
| `bundle-stages-firefox-embed` | StageFirefoxExtensionEmbed (or Bundle dual-stage) writes `browseragent/embedded/extension-firefox` with `manifest.json` under temp Root |
| `chrome-regression/build-extension-shell-chrome` | `BuildExtensionShell` still targets `Chrome-Ext-Browser-Agent/build` |

**Leaf count: 5**

## How to Run

```sh
doctest vet ./tests/browser-agent-firefox-connect
doctest test ./tests/browser-agent-firefox-connect
# expect RED until Phase 1 implementer lands Firefox connect + embed stage
```

Module: `github.com/xhd2015/browser-agent`. Package under test: `browseragent`.

### Implementer contract (authoritative for GREEN)

#### Extension sources (`Firefox-Ext-Browser-Agent/public/`)

1. **contentScript.js**
   - Keep `window.__BROWSER_AGENT_EXT__` marker (product browser-agent; browser firefox OK).
   - Parse session from `/go?session=` (`URLSearchParams` / `location.search`) and/or
     `data-session-id` on documentElement.
   - `browser.runtime.sendMessage` **or** `chrome.runtime.sendMessage` with
     `{ type: "register", session_id, control_port, … }`.

2. **background.js**
   - Session map; on register open `WebSocket` to `/v1/ws?session=<id>` (or build WS URL
     with `?session=` and `/v1/ws`).
   - Send **hello** (`type: "hello"`) with version, features including `browser-agent`,
     optional `bundle_md5`, **`browser_product`** of firefox (`"firefox"` / `"Firefox"`).
   - Handle **ping→pong**; reconnect dead sockets (alarm and/or backoff); **keepalive**
     while WS open (interval ping / `keepaliveTimer` / equivalent). **Not** only an empty
     no-op 1-minute alarm body.

3. **manifest** — permissions for tabs/alarms/storage; host perms for localhost WS/HTTP
   (already present in P1 scaffold; not re-asserted here beyond source needs).

#### Go bundle (E1)

```go
// Preferred dedicated API:
const DefaultEmbedFirefoxExtensionRel = "browseragent/embedded/extension-firefox"

// StageFirefoxExtensionEmbed stages Firefox public→build (or fixture) into
// {Root}/browseragent/embedded/extension-firefox (or EmbedFirefoxExtensionRel)
// and returns the absolute embed directory. Must write manifest.json.
func StageFirefoxExtensionEmbed(opts FirefoxEmbedOptions) (embedDir string, err error)

// FirefoxEmbedOptions fields (suggested):
//   Root, UseFixture, FixtureFirefoxExtensionDir, EmbedFirefoxExtensionRel

// Acceptable equivalent observed by this tree:
//   Bundle also populates {Root}/browseragent/embedded/extension-firefox
//   when Firefox sources or fixtures/extension-firefox exist.
//
// This tree's Run:
//   1) BuildFirefoxExtensionShell(ShellRoot) when Firefox public is staged under ShellRoot
//   2) StageFirefoxExtensionEmbed when exported (see browseragent.StageFirefoxExtensionEmbed)
//   3) Bundle(UseFixture) so dual-stage Bundle implementers also go GREEN
//   4) Assert filesystem: embed path has manifest.json (prefer gecko.id / Firefox identity)
```

**Behavior:**

1. Real sources: `BuildFirefoxExtensionShell(root)` then stage `build/` → embed rel.
2. Fixture: copy `browseragent/fixtures/extension-firefox` (or override) → embed rel.
3. Temp Root only in tests — do not require live ModuleRoot mutation.
4. No real Firefox / network / npm.

Chrome: `BuildExtensionShell` remains `Chrome-Ext-Browser-Agent` only.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface under test.
const (
	ModeExtSource        = "ext-source"
	ModeBundleFirefox    = "bundle-stages-firefox-embed"
	ModeChromeRegression = "chrome-regression"
)

// ExtSourceTarget for ModeExtSource.
const (
	ExtSrcContentScriptRegister = "content-script-register"
	ExtSrcBackgroundHelloWS     = "background-hello-ws"
	ExtSrcBackgroundKeepalive   = "background-keepalive"
)

// ChromeRegressionOp for ModeChromeRegression.
const (
	ChromeRegressionOpBuildShellChrome = "build-extension-shell-chrome"
)

// Default firefox embed rel (mirrors implementer DefaultEmbedFirefoxExtensionRel).
const defaultEmbedFirefoxExtensionRel = "browseragent/embedded/extension-firefox"

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string

	// ext-source
	ExtSourceTarget string

	// bundle-stages-firefox-embed
	BundleRoot                 string
	UseFixture                 bool
	FixtureFirefoxExtensionDir string
	FixtureSessionPageDir      string
	// ShellRoot: temp root with Firefox-Ext-Browser-Agent/public for real-source stage.
	// Also used by chrome-regression for Chrome-Ext public fixture.
	ShellRoot string

	// chrome-regression
	ChromeRegressionOp string
}

// Response holds probe outcomes for all modes.
type Response struct {
	// ext-source
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	ErrText      string

	// bundle-stages-firefox-embed
	FirefoxEmbedDir     string
	FirefoxManifestPath string
	FirefoxManifestText string
	FirefoxManifestOK   bool
	UsedFixture         bool
	BundleErr           string
	BuildShellDir       string
	StageAPIUsed        string // "Bundle" | "BuildFirefox+probe" | "none"

	// chrome-regression
	ChromeBuildDir string
	BuildDir       string
	BuildManifest  string
	BuildShellErr  string
}

// Run executes the scenario selected by req.Mode and leaf Setup narrowing.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		t.Fatal("req is nil")
	}
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}
	switch req.Mode {
	case ModeExtSource:
		return runExtSource(t, req)
	case ModeBundleFirefox:
		return runBundleStagesFirefox(t, req)
	case ModeChromeRegression:
		return runChromeRegression(t, req)
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
	case ExtSrcContentScriptRegister:
		candidates := firefoxContentScriptCandidates(root)
		path, data, ok := firstExistingFile(candidates)
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "contentScript.js not found under Firefox-Ext-Browser-Agent"
		}
		return resp, nil

	case ExtSrcBackgroundHelloWS, ExtSrcBackgroundKeepalive:
		candidates := firefoxBackgroundCandidates(root)
		path, data, ok := firstExistingFile(candidates)
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "background.js not found under Firefox-Ext-Browser-Agent"
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown ExtSourceTarget %q", req.ExtSourceTarget)
	}
}

// runBundleStagesFirefox exercises the Phase-1 E1 staging contract under a temp root.
//
// GREEN when implementer makes either of these true after this Run:
//   1) Bundle dual-stages {Root}/browseragent/embedded/extension-firefox/manifest.json
//      (optionally via StageFirefoxExtensionEmbed called from Bundle), or
//   2) BuildFirefoxExtensionShell + stage path writes that embed dir for real sources.
//
// This Run:
//   - BuildFirefoxExtensionShell(ShellRoot) when Firefox public is present
//   - Bundle(UseFixture) with chrome+session fixtures under BundleRoot
//   - Probes expected firefox embed path(s) for manifest.json
// Transport error is nil so Assert owns the RED (missing embed) vs GREEN outcome.
func runBundleStagesFirefox(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BundleRoot == "" {
		t.Fatal("BundleRoot must be set for bundle-stages-firefox-embed")
	}
	resp := &Response{StageAPIUsed: "none"}

	stageRoot := req.BundleRoot
	if strings.TrimSpace(req.ShellRoot) != "" {
		stageRoot = req.ShellRoot
	}

	// Real-source path: BuildFirefoxExtensionShell when public/manifest exists under stageRoot.
	// Implementer StageFirefoxExtensionEmbed (or Bundle) should then copy build/ → embed.
	publicMani := filepath.Join(stageRoot, "Firefox-Ext-Browser-Agent", "public", "manifest.json")
	if fileExists(publicMani) {
		buildDir, berr := browseragent.BuildFirefoxExtensionShell(stageRoot)
		if berr != nil {
			resp.BundleErr = berr.Error()
		} else {
			resp.BuildShellDir = buildDir
			resp.StageAPIUsed = "BuildFirefox+probe"
		}
	}

	// Bundle(UseFixture): implementer may dual-stage firefox embed as a side effect.
	if req.UseFixture && req.FixtureSessionPageDir != "" {
		if err := runBundleUseFixture(t, req, resp); err != nil && resp.BundleErr == "" {
			resp.BundleErr = err.Error()
		}
	}

	// Probe expected embed locations (stageRoot and/or BundleRoot).
	candidates := []string{
		filepath.Join(stageRoot, filepath.FromSlash(defaultEmbedFirefoxExtensionRel)),
		filepath.Join(req.BundleRoot, filepath.FromSlash(defaultEmbedFirefoxExtensionRel)),
	}
	for _, dir := range candidates {
		if fileExists(filepath.Join(dir, "manifest.json")) {
			resp.FirefoxEmbedDir = dir
			if resp.StageAPIUsed == "none" || resp.StageAPIUsed == "BuildFirefox+probe" {
				// Bundle populated embed (or Stage called from Bundle).
				if strings.Contains(filepath.ToSlash(dir), "embedded/extension-firefox") {
					if fileExists(filepath.Join(req.BundleRoot, filepath.FromSlash(defaultEmbedFirefoxExtensionRel), "manifest.json")) &&
						dir == filepath.Join(req.BundleRoot, filepath.FromSlash(defaultEmbedFirefoxExtensionRel)) {
						resp.StageAPIUsed = "Bundle"
					}
				}
			}
			break
		}
	}
	if resp.FirefoxEmbedDir == "" {
		resp.FirefoxEmbedDir = filepath.Join(stageRoot, filepath.FromSlash(defaultEmbedFirefoxExtensionRel))
	}
	fillFirefoxEmbedProbe(resp)
	return resp, nil
}

func runBundleUseFixture(t *testing.T, req *Request, resp *Response) error {
	t.Helper()
	// Primary chrome extension slot: prefer chrome fixture so Bundle succeeds while
	// firefox embed staging is still missing (RED). Implementer dual-stages firefox
	// from fixtures/extension-firefox or ShellRoot public.
	extFix := filepath.Join(req.ModuleRoot, "browseragent", "fixtures", "extension")
	if !fileExists(filepath.Join(extFix, "manifest.json")) {
		extFix = req.FixtureFirefoxExtensionDir
	}
	opts := browseragent.BundleOptions{
		Root:                  req.BundleRoot,
		UseFixture:            true,
		FixtureExtensionDir:   extFix,
		FixtureSessionPageDir: req.FixtureSessionPageDir,
	}
	// Point fixture discovery at ModuleRoot copies under BundleRoot when staged by leaf Setup.
	if req.FixtureFirefoxExtensionDir != "" {
		// Ensure firefox fixture is visible under BundleRoot for implementer discovery.
		// (Leaf Setup copies fixtures/extension-firefox into BundleRoot when UseFixture.)
		_ = req.FixtureFirefoxExtensionDir
	}
	res, err := browseragent.Bundle(opts)
	if err != nil {
		return err
	}
	if res != nil {
		resp.UsedFixture = res.UsedFixture || resp.UsedFixture
	}
	ff := filepath.Join(req.BundleRoot, filepath.FromSlash(defaultEmbedFirefoxExtensionRel))
	if fileExists(filepath.Join(ff, "manifest.json")) {
		resp.FirefoxEmbedDir = ff
		resp.StageAPIUsed = "Bundle"
	}
	return nil
}

func fillFirefoxEmbedProbe(resp *Response) {
	if resp.FirefoxEmbedDir == "" {
		return
	}
	mp := filepath.Join(resp.FirefoxEmbedDir, "manifest.json")
	resp.FirefoxManifestPath = mp
	data, err := os.ReadFile(mp)
	if err != nil {
		resp.FirefoxManifestOK = false
		return
	}
	resp.FirefoxManifestOK = true
	resp.FirefoxManifestText = string(data)
}

func runChromeRegression(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ChromeRegressionOp == "" {
		t.Fatal("ChromeRegressionOp must be set")
	}
	resp := &Response{}
	switch req.ChromeRegressionOp {
	case ChromeRegressionOpBuildShellChrome:
		root := req.ShellRoot
		if root == "" {
			t.Fatal("ShellRoot must be set for build-extension-shell-chrome")
		}
		buildDir, err := browseragent.BuildExtensionShell(root)
		if err != nil {
			resp.BuildShellErr = err.Error()
			return resp, err
		}
		resp.ChromeBuildDir = buildDir
		resp.BuildDir = buildDir
		resp.BuildManifest = filepath.Join(buildDir, "manifest.json")
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown ChromeRegressionOp %q", req.ChromeRegressionOp)
	}
}

func firefoxBackgroundCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "public", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "build", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "background.js"),
	}
}

func firefoxContentScriptCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "public", "contentScript.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "public", "content.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "build", "contentScript.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "contentScript.js"),
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

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
```
