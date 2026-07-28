# browser-agent Firefox core jobs (Phase 2 B)

Classic TDD for **Phase 2** of Firefox support: implement **core job handlers** in
`Firefox-Ext-Browser-Agent` background **without** `chrome.debugger` / full CDP.

| Surface | What is under test |
|---------|-------------------|
| handleJob dispatch | Real branches for core job types (not Phase-1 stub-only) |
| `info` | Tabs list shape (`handleInfoJob` or equivalent) |
| `create_tab` | `browser.tabs.create` / `tabs.create` |
| `eval` / `run` | `scripting.executeScript` or `tabs.executeScript` |
| `screenshot` | `tabs.captureVisibleTab` (viewport) |
| `logs` | Best-effort limited or empty entries OK with `type: "logs"` |
| Anti-stub | Core jobs are **not** only `"not implemented"` Phase-1 stubs |
| Manifest | `scripting` (when needed) + host permissions for page injection |

**No real Firefox. No network. No CDP.** Source markers + manifest FS only.

**Classic TDD — RED** until implementer ports non-debugger job paths into
`Firefox-Ext-Browser-Agent/public/background.js` (+ manifest) and syncs build/embed.

### Out of scope (later phases)

- Full CDP matrix / `chrome.debugger` (Phase 3)
- SPA Firefox install UX (Phase 4)
- Chrome path changes

## Version

0.0.2

# DSN (Domain Specific Notion)

**Control Server** sends **job** envelopes over the per-session **WebSocket** already
established in Phase 1. The **Firefox Background Worker** receives `type: "job"`,
parses `payload.type` / `job_type`, and runs a **job handler** that uses **WebExtensions
APIs** (not debugger CDP):

```text
Control Server
  -> WS job { type: info|eval|run|logs|screenshot|create_tab, params, tab_id? }
  -> Background handleJob
       case info        -> list session-window tabs; return tabs[] shape
       case create_tab  -> browser.tabs.create({ windowId, url?, active? })
       case eval|run    -> scripting.executeScript OR tabs.executeScript
       case screenshot  -> tabs.captureVisibleTab (viewport png/jpeg)
       case logs        -> best-effort entries[] (may be empty) + type logs
  -> WS result { ok, data, error }
```

**Phase-1 stub (must leave):** a single `handleJob` that always
`sendJobResult(..., stub:true, "not implemented: firefox job runner…")` is
**insufficient** for Phase 2 core types (`info`, `eval`, `create_tab`, `screenshot`).

**Manifest:** when using `browser.scripting.executeScript`, add **`scripting`**
permission; broaden **host_permissions** so page injection/capture can reach
user tabs (localhost-only hosts are not enough for arbitrary page jobs).

**Test Client** reads `Firefox-Ext-Browser-Agent` sources (public preferred) and
asserts tokens — never launches a browser.

```text
Test Client
  -> read public/background.js (or build/)
  -> assert handleJob branches + API markers
  -> read public/manifest.json
  -> assert scripting / host_permissions for injection
```

## Decision Tree

```
browser-agent-firefox-jobs
├── background-source/                           [Firefox-Ext public/build background.js]
│   ├── handle-job-dispatch/                       real case/branch tokens for core types
│   ├── info-tabs-shape/                           handleInfoJob / info + tabs list shape
│   ├── create-tab/                                create_tab + tabs.create
│   ├── eval-run-execute-script/                   eval + run + executeScript
│   ├── screenshot-capture-visible/                screenshot + captureVisibleTab
│   ├── logs-best-effort/                          logs type; empty entries OK
│   └── core-jobs-not-stub-only/                   explicit: not Phase-1 not-implemented only
└── manifest/                                    [Firefox-Ext manifest.json]
    └── scripting-and-host-permissions/            scripting and/or injection host perms
```

### Parameter significance (high → low)

1. **Surface / Mode** — background source vs manifest (different files / Run branches).
2. **Within background-source** — which job contract or anti-stub property
   (MECE: dispatch vs each core job vs logs vs global anti-stub).
3. **Within manifest** — permissions for scripting/injection hosts.

## Test Index

| Leaf | Scenario |
|------|----------|
| `background-source/handle-job-dispatch` | `background.js` dispatches job types with real branch tokens: `info`, `eval`, `run`, `screenshot`, `create_tab`, `logs` (quoted/`case` forms preferred) |
| `background-source/info-tabs-shape` | info path: `handleInfoJob` name **or** `"info"` branch; tabs listing API; result shape mentions `tabs` and tab fields (`id`/`url`/`title`) |
| `background-source/create-tab` | `"create_tab"` job token + `tabs.create` (`browser.tabs.create` / `chrome.tabs.create` / `.tabs.create`) |
| `background-source/eval-run-execute-script` | `"eval"` + `"run"` tokens + `executeScript` (`scripting.executeScript` or `tabs.executeScript`) |
| `background-source/screenshot-capture-visible` | `"screenshot"` token + `captureVisibleTab` |
| `background-source/logs-best-effort` | `"logs"` branch; may return empty `entries`; must not require debugger-only Log.enable |
| `background-source/core-jobs-not-stub-only` | For `info`, `eval`, `create_tab`, `screenshot`: real branches + APIs; **fails** if handleJob is only Phase-1 `"not implemented"` stub |
| `manifest/scripting-and-host-permissions` | `manifest.json`: if scripting path used, `scripting` permission; host_permissions broader than localhost-only **or** explicit inject/capture-ready hosts |

**Leaf count: 8**

## How to Run

```sh
doctest vet ./tests/browser-agent-firefox-jobs
doctest test ./tests/browser-agent-firefox-jobs
# expect RED until Phase 2 implementer lands Firefox core job handlers + manifest
```

Module: `github.com/xhd2015/browser-agent`. Under test: `Firefox-Ext-Browser-Agent`
sources (read-only FS). No package Go job-runner changes required for this tree.

### Implementer contract (authoritative for GREEN)

#### `Firefox-Ext-Browser-Agent/public/background.js`

1. **`handleJob`** — switch/if on `jobType` with real handlers for at least:
   `info`, `eval`, `run`, `logs`, `screenshot`, `create_tab`.
   Must **not** be only Phase-1:
   `sendJobResult(..., { stub: true }, "not implemented: firefox job runner…")`
   for those core types.

2. **`info`** — prefer `handleInfoJob` (name optional but preferred marker). List
   capturable/session-window tabs (`browser.tabs.query` or shared helper). Return
   data including a **`tabs`** array with elements carrying **`id`**, **`url`**,
   **`title`** (index/active/role optional, Chrome-parity welcome).

3. **`create_tab`** — `browser.tabs.create` (or `chrome.tabs.create` polyfill /
   `tabs.create`) scoped to session `windowId` when bound; result includes
   `type: "create_tab"` and `tab_id`.

4. **`eval` / `run`** — inject expression/source via
   **`browser.scripting.executeScript`** **or** **`browser.tabs.executeScript`**
   (or chrome.* equivalent). No `chrome.debugger` / `Runtime.evaluate` required
   in Phase 2.

5. **`screenshot`** — **`tabs.captureVisibleTab`** (viewport). Return base64-ish
   image data + `type: "screenshot"`. Full-page CDP beyond viewport is Phase 3.

6. **`logs`** — best-effort: return `{ type: "logs", entries: [...] }` with
   empty or limited buffer OK. Debugger `Log.enable` not required.

#### `Firefox-Ext-Browser-Agent/public/manifest.json`

7. Permissions:
   - Keep `tabs`, `alarms`, `storage` as needed.
   - Add **`scripting`** when using `scripting.executeScript`.
   - **host_permissions** must allow injection/capture on pages the agent
     targets — localhost-only is insufficient if jobs run on arbitrary http(s)
     tabs; prefer adding broader hosts (e.g. `http://*/*`, `https://*/*`, or
     `<all_urls>`) as appropriate for temporary add-on use.

8. Sync **build/** and **browseragent/embedded/extension-firefox** (and fixtures
   if used) after public changes (same practice as Phase 1).

**Not required this phase:** CDP job type, `chrome.debugger`, SPA install panel.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mode — top-level surface under test.
const (
	ModeBackgroundSource = "background-source"
	ModeManifest         = "manifest"
)

// BackgroundSourceTarget for ModeBackgroundSource.
const (
	BgSrcHandleJobDispatch       = "handle-job-dispatch"
	BgSrcInfoTabsShape           = "info-tabs-shape"
	BgSrcCreateTab               = "create-tab"
	BgSrcEvalRunExecuteScript    = "eval-run-execute-script"
	BgSrcScreenshotCaptureVisible = "screenshot-capture-visible"
	BgSrcLogsBestEffort          = "logs-best-effort"
	BgSrcCoreJobsNotStubOnly     = "core-jobs-not-stub-only"
)

// ManifestTarget for ModeManifest.
const (
	ManifestScriptingAndHosts = "scripting-and-host-permissions"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string

	// background-source
	BackgroundSourceTarget string

	// manifest
	ManifestTarget string
}

// Response holds probe outcomes for all modes.
type Response struct {
	// shared FS probe
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	ErrText      string
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
	case ModeBackgroundSource:
		return runBackgroundSource(t, req)
	case ModeManifest:
		return runManifest(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runBackgroundSource(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BackgroundSourceTarget == "" {
		t.Fatal("BackgroundSourceTarget must be set")
	}
	// All background leaves share the same file probe; Assert specializes.
	_ = req.BackgroundSourceTarget
	resp := &Response{FileContents: map[string]string{}}
	candidates := firefoxBackgroundCandidates(req.ModuleRoot)
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
}

func runManifest(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ManifestTarget == "" {
		t.Fatal("ManifestTarget must be set")
	}
	_ = req.ManifestTarget
	resp := &Response{FileContents: map[string]string{}}
	candidates := firefoxManifestCandidates(req.ModuleRoot)
	path, data, ok := firstExistingFile(candidates)
	resp.FileExists = ok
	if ok {
		resp.FoundPaths = []string{path}
		resp.FileContents[path] = string(data)
		resp.CombinedText = string(data)
	} else {
		resp.ErrText = "manifest.json not found under Firefox-Ext-Browser-Agent"
	}
	return resp, nil
}

func firefoxBackgroundCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "public", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "build", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "background.js"),
	}
}

func firefoxManifestCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "public", "manifest.json"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "build", "manifest.json"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "manifest.json"),
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

// jobTypeTokenPresent accepts common JS job-type branch forms.
func jobTypeTokenPresent(src, jt string) bool {
	candidates := []string{
		`"` + jt + `"`,
		`'` + jt + `'`,
		"`" + jt + "`",
		`case "` + jt + `"`,
		`case '` + jt + `'`,
		`type === "` + jt + `"`,
		`jobType === "` + jt + `"`,
		`job_type === "` + jt + `"`,
		`=== "` + jt + `"`,
		`== "` + jt + `"`,
	}
	for _, c := range candidates {
		if strings.Contains(src, c) {
			return true
		}
	}
	// Longer unique type names: bare token OK.
	if jt == "screenshot" || jt == "create_tab" || jt == "logs" {
		return strings.Contains(src, jt)
	}
	return false
}

// hasTabsCreate reports tabs.create usage (browser/chrome polyfill OK).
func hasTabsCreate(src string) bool {
	return strings.Contains(src, "tabs.create") ||
		strings.Contains(src, "browser.tabs.create") ||
		strings.Contains(src, "chrome.tabs.create")
}

// hasExecuteScript reports scripting or tabs executeScript (Phase 2 eval/run path).
func hasExecuteScript(src string) bool {
	low := strings.ToLower(src)
	if strings.Contains(src, "executeScript") {
		return true
	}
	// Tolerate spaced / dotted forms already covered by executeScript.
	return strings.Contains(low, "scripting.executescript") ||
		strings.Contains(low, "tabs.executescript")
}

// hasCaptureVisibleTab reports viewport screenshot API.
func hasCaptureVisibleTab(src string) bool {
	return strings.Contains(src, "captureVisibleTab") ||
		strings.Contains(strings.ToLower(src), "capturevisibletab")
}

// hasTabsQueryOrList reports tab listing used by info.
func hasTabsQueryOrList(src string) bool {
	low := strings.ToLower(src)
	if strings.Contains(src, "tabs.query") || strings.Contains(src, "tabs.get") {
		return true
	}
	if strings.Contains(low, "listcapturable") || strings.Contains(low, "list_tabs") {
		return true
	}
	if strings.Contains(low, "querytabs") || strings.Contains(low, "listtabs") {
		return true
	}
	// "tabs" array construction near info handler is weak; prefer query.
	return strings.Contains(src, "browser.tabs") && strings.Contains(low, "query")
}

// isPhase1JobStubOnly detects Phase-1 connect-only stub handleJob.
// GREEN when real core job branches + APIs are present even if a fallback
// "not implemented" remains for unknown types / CDP.
func isPhase1JobStubOnly(src string) bool {
	low := strings.ToLower(src)
	hasStubMsg := strings.Contains(low, "not implemented") &&
		(strings.Contains(low, "phase 1") ||
			strings.Contains(low, "firefox job runner") ||
			strings.Contains(low, "connect only") ||
			strings.Contains(low, "stub: true") ||
			strings.Contains(src, "stub: true") ||
			strings.Contains(src, `"stub": true`) ||
			strings.Contains(src, "stub:true"))

	// Real dispatch for the four mandatory Phase-2 core types.
	coreOK := jobTypeTokenPresent(src, "info") &&
		jobTypeTokenPresent(src, "eval") &&
		jobTypeTokenPresent(src, "create_tab") &&
		jobTypeTokenPresent(src, "screenshot")

	apisOK := hasExecuteScript(src) && hasCaptureVisibleTab(src) && hasTabsCreate(src) &&
		(hasTabsQueryOrList(src) || strings.Contains(src, "handleInfoJob") ||
			strings.Contains(low, "handleinfojob"))

	if coreOK && apisOK {
		return false
	}
	// Stub message without real core implementation.
	if hasStubMsg {
		return true
	}
	// No stub message but also no real core handlers → still "stub only" for Phase 2.
	return !coreOK || !apisOK
}
```
