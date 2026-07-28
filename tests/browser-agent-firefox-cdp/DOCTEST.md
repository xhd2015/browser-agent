# browser-agent Firefox CDP matrix (Phase 3 C)

Classic TDD for **Phase 3** of Firefox support: implement a **partial CDP job
matrix** in `Firefox-Ext-Browser-Agent` background — map a small set of CDP
methods onto WebExtensions APIs (no real `chrome.debugger`, no full Network/DOM
CDP).

| Surface | What is under test |
|---------|-------------------|
| handleJob `cdp` case | Dedicated `case "cdp"` / branch (not only default unknown) |
| `Page.navigate` | Maps to `tabs.update` (url) |
| `Runtime.evaluate` | Maps to / reuses eval path (`executeScript` / `handleEvalJob`) |
| Unsupported methods | Clear error containing **not supported** and **firefox** (or similar) |
| Anti-generic | `cdp` is **not** only the generic unknown job-type path |
| Optional soft | `Target.createTarget` → create_tab path (soft polyfill) |

**No real Firefox. No network. No debugger attach.** Source markers on
`Firefox-Ext-Browser-Agent` background only.

**Classic TDD — RED** until implementer lands a dedicated `cdp` handler with the
Phase 3 method matrix.

### Out of scope

- Full Network / DOM CDP surface
- SPA Firefox install UX (Phase 4)
- Chrome path changes
- Real Firefox browser

## Version

0.0.2

# DSN (Domain Specific Notion)

**Control Server** posts **job** envelopes with `type: "cdp"` over the per-session
**WebSocket**. The **Firefox Background Worker** receives the job, enters a
**dedicated `cdp` branch** of **handleJob**, and routes `params.method` through a
**CDP matrix**:

```text
Control Server
  -> WS job { type: cdp, params: { method, params }, tab_id? }
  -> Background handleJob
       case "cdp" -> handleCdpJob (or equivalent)
            method Page.navigate
              -> browser.tabs.update(tabId, { url })
            method Runtime.evaluate
              -> reuse eval path (executeScript / handleEvalJob)
            method Target.createTarget  (optional soft)
              -> create_tab / tabs.create path
            other methods
              -> clear error: "…not supported…firefox…" (or similar)
  -> WS result { ok, data, error }
```

**Phase 2 default (must leave for `cdp`):** treating `cdp` only as unknown job
type via `default:` with a generic
`"not implemented: firefox job type=" + jobType + " (phase 2…)"` is
**insufficient**. Phase 3 requires an explicit `cdp` case and method routing.

**No chrome.debugger:** Firefox cannot attach the Chrome debugger API. Supported
methods are **polyfilled** with tabs/scripting APIs; unsupported methods must
fail with a **product-owned** message (not a silent fallthrough).

**Test Client** reads `Firefox-Ext-Browser-Agent` sources (public preferred) and
asserts tokens — never launches a browser.

```text
Test Client
  -> read public/background.js (or build/)
  -> assert case "cdp", Page.navigate→tabs.update, Runtime.evaluate→eval path
  -> assert unsupported message tokens (not supported + firefox)
  -> assert cdp is not only generic unknown-job default
```

## Decision Tree

```
browser-agent-firefox-cdp
└── background-source/                              [Firefox-Ext public/build background.js]
    ├── handle-job-cdp-case/                          dedicated case "cdp" / branch token
    ├── page-navigate-tabs-update/                    Page.navigate + tabs.update
    ├── runtime-evaluate-eval-path/                   Runtime.evaluate + eval/executeScript reuse
    ├── unsupported-method-message/                   "not supported" + "firefox" (or similar)
    ├── cdp-not-generic-unknown-only/                 anti: not only default unknown job type
    └── target-create-target-soft/                    soft: Target.createTarget → create_tab
```

### Parameter significance (high → low)

1. **Surface / Mode** — background-source only (single file family; all leaves share probe).
2. **CDP matrix aspect** — dispatch case vs supported method map vs unsupported
   error vs anti-generic vs optional Target soft polyfill (MECE branches).

## Test Index

| Leaf | Scenario |
|------|----------|
| `background-source/handle-job-cdp-case` | `background.js` has dedicated job-type branch for **`cdp`** (`case "cdp"` / quoted token / `jobType === "cdp"`); prefer `handleCdpJob` name |
| `background-source/page-navigate-tabs-update` | **`Page.navigate`** token present; navigation via **`tabs.update`** (browser/chrome OK) |
| `background-source/runtime-evaluate-eval-path` | **`Runtime.evaluate`** token present; reuses eval path (`handleEvalJob` and/or `executeScript`) |
| `background-source/unsupported-method-message` | Unsupported CDP methods surface a clear error containing **not supported** (or **unsupported**) and **firefox** (case-insensitive OK) |
| `background-source/cdp-not-generic-unknown-only` | Explicit: `cdp` must **not** be handled only by Phase-2 generic unknown default; dedicated cdp case + method routing markers required |
| `background-source/target-create-target-soft` | Soft: if `Target.createTarget` appears, must also touch create_tab / `tabs.create` path; absence of Target polyfill alone does **not** fail (optional) |

**Leaf count: 6**

## How to Run

```sh
doctest vet ./tests/browser-agent-firefox-cdp
doctest test ./tests/browser-agent-firefox-cdp
# expect RED until Phase 3 implementer lands Firefox cdp job matrix
```

Module: `github.com/xhd2015/browser-agent`. Under test: `Firefox-Ext-Browser-Agent`
sources (read-only FS). No package Go job-runner changes required for this tree.

### Implementer contract (authoritative for GREEN)

#### `Firefox-Ext-Browser-Agent/public/background.js`

1. **`handleJob` — `case "cdp"`** (or equivalent branch) must exist. Must **not**
   leave `cdp` solely to the Phase-2 `default:` unknown-type stub.

2. **`Page.navigate`** — when `params.method` (or `cdp_method` / `cdpMethod`) is
   `Page.navigate`, navigate the target tab via
   **`browser.tabs.update` / `chrome.tabs.update` / `tabs.update`** with the URL
   from `params.params.url` (or equivalent). Do **not** require `chrome.debugger`.

3. **`Runtime.evaluate`** — when method is `Runtime.evaluate`, evaluate expression
   by **reusing the eval path**: call `handleEvalJob` (preferred) and/or
   `scripting.executeScript` / `tabs.executeScript` with the expression from
   CDP params (`expression` / similar). Return a cdp-shaped result (method +
   type `"cdp"`) is welcome but not asserted here beyond markers.

4. **Unsupported methods** — any other CDP method must fail with a **clear
   product error** whose message contains both:
   - **`not supported`** or **`unsupported`** (case-insensitive), and
   - **`firefox`** (case-insensitive),
   e.g. `"CDP method X is not supported on firefox"` or similar.

5. **Optional soft:** `Target.createTarget` may map to the existing
   `create_tab` / `tabs.create` path. Soft leaf only asserts coupling when the
   Target token is present.

6. Sync **build/** and **browseragent/embedded/extension-firefox** (and fixtures
   if used) after public changes (same practice as Phase 1–2).

**Not required this phase:** Network.*, DOM.*, full Page.* beyond navigate,
`chrome.debugger`, SPA install panel.

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
)

// BackgroundSourceTarget for ModeBackgroundSource.
const (
	BgSrcHandleJobCdpCase            = "handle-job-cdp-case"
	BgSrcPageNavigateTabsUpdate      = "page-navigate-tabs-update"
	BgSrcRuntimeEvaluateEvalPath     = "runtime-evaluate-eval-path"
	BgSrcUnsupportedMethodMessage    = "unsupported-method-message"
	BgSrcCdpNotGenericUnknownOnly    = "cdp-not-generic-unknown-only"
	BgSrcTargetCreateTargetSoft      = "target-create-target-soft"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string

	// background-source
	BackgroundSourceTarget string
}

// Response holds probe outcomes.
type Response struct {
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

func firefoxBackgroundCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "public", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "build", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "background.js"),
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
	return false
}

// hasDedicatedCdpCase reports an explicit cdp job branch (not bare substring only).
func hasDedicatedCdpCase(src string) bool {
	if jobTypeTokenPresent(src, "cdp") {
		return true
	}
	// Named handler is strong evidence even without quoted case form.
	if strings.Contains(src, "handleCdpJob") || strings.Contains(src, "handleCDPJob") {
		return true
	}
	return false
}

// hasTabsUpdate reports tabs.update usage (browser/chrome polyfill OK).
func hasTabsUpdate(src string) bool {
	return strings.Contains(src, "tabs.update") ||
		strings.Contains(src, "browser.tabs.update") ||
		strings.Contains(src, "chrome.tabs.update")
}

// hasExecuteScript reports scripting or tabs executeScript.
func hasExecuteScript(src string) bool {
	if strings.Contains(src, "executeScript") {
		return true
	}
	low := strings.ToLower(src)
	return strings.Contains(low, "scripting.executescript") ||
		strings.Contains(low, "tabs.executescript")
}

// hasEvalPathReuse reports that a Runtime.evaluate CDP branch can reuse the
// Phase-2 eval path. Only meaningful once hasRuntimeEvaluate is true.
func hasEvalPathReuse(src string) bool {
	if !hasRuntimeEvaluate(src) {
		return false
	}
	if strings.Contains(src, "handleEvalJob") {
		return true
	}
	if hasExecuteScript(src) {
		return true
	}
	low := strings.ToLower(src)
	return strings.Contains(low, "handleeval") || strings.Contains(low, "executescriptintab")
}

// hasUnsupportedFirefoxMessage reports product-owned unsupported CDP wording.
func hasUnsupportedFirefoxMessage(src string) bool {
	low := strings.ToLower(src)
	hasUnsup := strings.Contains(low, "not supported") ||
		strings.Contains(low, "unsupported")
	hasFirefox := strings.Contains(low, "firefox")
	return hasUnsup && hasFirefox
}

// cdpMethodTokenPresent requires method as a branch/compare string, not a
// prose comment like "No chrome.debugger / Runtime.evaluate".
func cdpMethodTokenPresent(src, method string) bool {
	candidates := []string{
		`"` + method + `"`,
		`'` + method + `'`,
		"`" + method + "`",
		`case "` + method + `"`,
		`case '` + method + `'`,
		`=== "` + method + `"`,
		`== "` + method + `"`,
		`=== '` + method + `'`,
		`== '` + method + `'`,
		`method === "` + method + `"`,
		`method == "` + method + `"`,
	}
	for _, c := range candidates {
		if strings.Contains(src, c) {
			return true
		}
	}
	return false
}

// hasPageNavigate reports Page.navigate as a CDP method branch token.
func hasPageNavigate(src string) bool {
	return cdpMethodTokenPresent(src, "Page.navigate") ||
		// bare Page.navigate only if clearly assigned/compared (avoid pure prose);
		// still accept unquoted in switch-like forms common in JS source.
		strings.Contains(src, "Page.navigate") &&
			(strings.Contains(src, "method") || strings.Contains(src, "cdp") ||
				strings.Contains(src, "handleCdp") || strings.Contains(src, "tabs.update"))
}

// hasRuntimeEvaluate reports Runtime.evaluate as a CDP method branch token.
// Quoted/compare forms only — a "No … Runtime.evaluate" comment must not match.
func hasRuntimeEvaluate(src string) bool {
	return cdpMethodTokenPresent(src, "Runtime.evaluate")
}

// isCdpOnlyGenericUnknown detects Phase-2 style: cdp falls through default
// unknown job type with no dedicated case/handler/method matrix.
//
// GREEN (returns false) when dedicated cdp branch + at least one supported
// method marker (Page.navigate or Runtime.evaluate) are present.
// RED (returns true) when cdp is missing as a case and only generic unknown
// messaging exists (or no cdp surface at all).
func isCdpOnlyGenericUnknown(src string) bool {
	dedicated := hasDedicatedCdpCase(src)
	matrix := hasPageNavigate(src) || hasRuntimeEvaluate(src)
	if dedicated && matrix {
		return false
	}
	// Explicit comments / default path calling out cdp as unimplemented unknown.
	low := strings.ToLower(src)
	genericUnknown := (strings.Contains(low, "unknown") && strings.Contains(low, "cdp")) ||
		(strings.Contains(low, "not implemented") && strings.Contains(low, "job type")) ||
		strings.Contains(low, "phase 2 core jobs only") ||
		(strings.Contains(low, "e.g. cdp") && strings.Contains(low, "unimplemented"))
	if !dedicated {
		// No dedicated case: always treat as generic-only for Phase 3.
		return true
	}
	// Dedicated case token exists but no method matrix → still incomplete.
	if !matrix {
		return true
	}
	_ = genericUnknown
	return false
}

// hasTargetCreateTarget reports Target.createTarget token.
func hasTargetCreateTarget(src string) bool {
	return strings.Contains(src, "Target.createTarget") ||
		strings.Contains(src, `"Target.createTarget"`) ||
		strings.Contains(src, `'Target.createTarget'`)
}

// hasCreateTabPath reports create_tab job or tabs.create API.
func hasCreateTabPath(src string) bool {
	if strings.Contains(src, "tabs.create") ||
		strings.Contains(src, "browser.tabs.create") ||
		strings.Contains(src, "chrome.tabs.create") {
		return true
	}
	return jobTypeTokenPresent(src, "create_tab") ||
		strings.Contains(src, "handleCreateTabJob") ||
		strings.Contains(src, "createTabInSession")
}
```
