# browser-agent session page dynamic browser detect (Phase 1)

Classic-TDD tree for **client-side install browser detection** on the session
page SPA (`react/src/ui/SessionPageApp.tsx` + helpers):

| Surface | What is under test |
|---------|-------------------|
| `detectRuntimeBrowser` | Pure UA helper: `Firefox/` → `firefox`, else `chrome` |
| `resolveInstallBrowser` | Priority: forced prop → `__BROWSER_AGENT_EXT__.browser` → UA Firefox → snap/boot/path → default chrome |
| `SessionPageApp` wiring | Uses helpers so Firefox UA / content-script marker win over chrome-stamped boot; pass result to `InstallGuideline` |

**Bug today:** `resolveInstallBrowser` prefers snap / boot / path only. Firefox
tabs with **chrome-stamped** meta/boot show Chrome install steps (no UA, no
`__BROWSER_AGENT_EXT__`).

**No real browser. No npm. No network.** Source / FS markers only.

## Mode

**Classic TDD** — desired priority + pure helpers are **not** implemented yet.
Current code has snap/boot/path-first resolution without UA / content-script
marker. Expect **RED** leaves until implementer lands Phase 1. Do **not**
implement production code from this tree.

### Plan phase

| Phase | Scope |
|-------|--------|
| **1 (this)** | Client detect for InstallGuideline — forced → `__BROWSER_AGENT_EXT__` → UA Firefox → snap/boot/path → chrome |
| **2 (later)** | Server Create stamps firefox extension path + browser in meta/boot |

### Out of scope Phase 1

- Fixing Create() stamping (Phase 2)
- Firefox connect/register bugs
- Changing InstallGuideline dual-path copy (keep both chrome + firefox paths)

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** opens the session page in **Firefox** (or Chrome). Before the
extension connects, the SPA must show the correct **InstallGuideline** steps.

Chrome-stamped session meta/boot is common today (Create path always chrome).
Detection must **not** trust boot alone when the live tab is Firefox.

```text
# Priority (highest → lowest) for install browser = chrome | firefox
1. forced prop  browser?: "chrome" | "firefox"   # explicit override
2. content-script marker  window.__BROWSER_AGENT_EXT__.browser
3. runtime UA  navigator.userAgent matches Firefox/  via detectRuntimeBrowser()
4. snap / boot / path  snap.browser, snap.browsers, installPath, __BROWSER_AGENT, #browser-agent-boot
5. default  chrome
```

**detectRuntimeBrowser**

```text
detectRuntimeBrowser(ua?)
  ua = ua ?? navigator.userAgent
  if ua contains "Firefox/" (case-sensitive token preferred) -> "firefox"
  else -> "chrome"
```

**resolveInstallBrowser**

```text
resolveInstallBrowser(forced, snap, installPath, …)
  if forced is chrome|firefox -> forced
  else if __BROWSER_AGENT_EXT__.browser is firefox -> firefox
  else if detectRuntimeBrowser() is firefox -> firefox
  else if snap/boot/path indicates firefox -> firefox
  else -> chrome
```

**SessionPageApp**

```text
SessionPageApp
  installBrowser = resolveInstallBrowser(...)
  -> InstallGuideline browser={installBrowser}
  -> Firefox UA + chrome boot still shows about:debugging install steps
```

**Test Client** reads module sources via **ModuleRoot** =
`filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))` — never launches a
browser.

## Decision Tree

```
browser-agent-session-page-browser-detect
├── detect-runtime-browser/                      [pure UA helper]
│   ├── ua-firefox/                                D1 Firefox/ → firefox
│   └── ua-non-firefox/                            D2 Chrome/empty/other → chrome
├── resolve-priority/                            [resolveInstallBrowser order]
│   ├── forced-prop-wins/                          P1 forced prop highest priority
│   ├── ext-marker-over-boot/                      P2 __BROWSER_AGENT_EXT__ over chrome boot
│   ├── ua-over-chrome-boot/                       P3 Firefox UA over chrome snap/boot/path
│   ├── snap-boot-path-fallback/                   P4 legacy snap/boot/path when higher absent
│   └── default-chrome/                            P5 no signals → chrome
└── session-page-wiring/                         [SessionPageApp integration]
    ├── helpers-named/                             W1 detectRuntimeBrowser + resolveInstallBrowser named
    ├── wires-install-browser/                     W2 resolve result → InstallGuideline browser=
    └── dual-paths-kept/                           W3 about:debugging + chrome://extensions still dual
```

### Parameter significance (high → low)

1. **Contract surface** — pure UA helper vs priority resolver vs SPA wiring
   (different behavioral units; same FS surface).
2. **Winning signal** (resolve only) — which input tier determines install browser
   (forced > EXT > UA > snap/boot/path > default).
3. **UA outcome** (detect only) — Firefox token vs non-Firefox default.
4. **Wiring detail** — helper names vs InstallGuideline prop vs dual-path keep.

## Test Index

| Leaf | Scenario |
|------|----------|
| `detect-runtime-browser/ua-firefox` | (D1) `detectRuntimeBrowser` present; Firefox UA (`Firefox/` token / equivalent) maps to `"firefox"` |
| `detect-runtime-browser/ua-non-firefox` | (D2) Non-Firefox / empty UA maps to `"chrome"` (default branch) |
| `resolve-priority/forced-prop-wins` | (P1) `resolveInstallBrowser` returns forced `chrome`\|`firefox` before EXT / UA / snap / boot |
| `resolve-priority/ext-marker-over-boot` | (P2) Reads `window.__BROWSER_AGENT_EXT__.browser`; EXT firefox wins over chrome-stamped snap/boot; checked **before** boot/path |
| `resolve-priority/ua-over-chrome-boot` | (P3) Uses `detectRuntimeBrowser` / UA Firefox **before** snap/boot/path so chrome-stamped boot does not force Chrome install steps |
| `resolve-priority/snap-boot-path-fallback` | (P4) When forced/EXT/FF-UA absent, snap `browser`/`browsers`, install path `browser-agent-firefox`, and/or boot still can yield firefox |
| `resolve-priority/default-chrome` | (P5) Default return is `"chrome"` when no firefox signal |
| `session-page-wiring/helpers-named` | (W1) Source defines both `detectRuntimeBrowser` and `resolveInstallBrowser` (export preferred; clear names required) |
| `session-page-wiring/wires-install-browser` | (W2) SessionPageApp calls `resolveInstallBrowser` and passes result into `InstallGuideline` `browser=` |
| `session-page-wiring/dual-paths-kept` | (W3) Combined SPA sources still carry Firefox (`about:debugging`) and Chrome (`chrome://extensions`) install paths |

**Leaf count: 10**

## How to Run

```sh
# module root
doctest vet ./tests/browser-agent-session-page-browser-detect
doctest test ./tests/browser-agent-session-page-browser-detect
# expect RED until Phase 1 implementer lands detect + priority
```

Module: `github.com/xhd2015/browser-agent`.  
Under test: `react/src/ui/SessionPageApp.tsx` (helpers may live there or a
sibling module imported by SessionPageApp), `react/src/ui/InstallGuideline.tsx`
(dual-path keep).

### Implementer contract (authoritative for GREEN)

#### `detectRuntimeBrowser`

```ts
// Prefer exported pure helper (name is a contract).
function detectRuntimeBrowser(userAgent?: string): "chrome" | "firefox"
// When userAgent omitted, read navigator.userAgent (guard typeof navigator).
// If UA includes "Firefox/" → "firefox"; else → "chrome".
```

#### `resolveInstallBrowser` priority (must not prefer chrome boot over UA/EXT)

```text
1. forced prop "chrome" | "firefox"  → return forced
2. window.__BROWSER_AGENT_EXT__.browser (lowercase) === "firefox" → firefox
3. detectRuntimeBrowser() === "firefox" → firefox
4. existing snap.browsers / snap.browser / installPath browser-agent-firefox /
   window.__BROWSER_AGENT.browser / #browser-agent-boot JSON → firefox when indicated
5. default → chrome
```

Do **not** let chrome-stamped boot/meta override steps 2–3.

#### `SessionPageApp`

```ts
const installBrowser = resolveInstallBrowser(browserProp, snap, installPath);
// ...
<InstallGuideline browser={installBrowser} ... />
```

Keep InstallGuideline dual paths (chrome + firefox). Phase 2 stamps server meta;
this phase is client-only.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Mode — all leaves share React source probe surface.
const ModeReactSrc = "react-src"

// ReactProbe identifies the leaf contract.
const (
	ReactProbeDetectUAFirefox          = "detect-ua-firefox"
	ReactProbeDetectUANonFirefox       = "detect-ua-non-firefox"
	ReactProbeResolveForcedProp        = "resolve-forced-prop"
	ReactProbeResolveExtOverBoot       = "resolve-ext-over-boot"
	ReactProbeResolveUAOverChromeBoot  = "resolve-ua-over-chrome-boot"
	ReactProbeResolveSnapBootPath      = "resolve-snap-boot-path"
	ReactProbeResolveDefaultChrome     = "resolve-default-chrome"
	ReactProbeWiringHelpersNamed       = "wiring-helpers-named"
	ReactProbeWiringInstallBrowser     = "wiring-install-browser"
	ReactProbeWiringDualPaths          = "wiring-dual-paths"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes (always ModeReactSrc for this tree).
	Mode string

	// ModuleRoot is module directory.
	// Root Setup sets from d.DOCTEST_ROOT/../..
	ModuleRoot string

	// ReactProbe selects which source contract Run/Assert exercise.
	ReactProbe string
}

// Response holds filesystem probe outcomes.
type Response struct {
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	// SessionPageText is SessionPageApp (+ helper modules) only.
	SessionPageText string
	// GuidelineText is InstallGuideline source when loaded.
	GuidelineText string
	ErrText       string
}

// Run loads SessionPageApp (+ helpers) and optionally InstallGuideline.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		t.Fatal("req is nil")
	}
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.Mode != ModeReactSrc {
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
	if req.ReactProbe == "" {
		t.Fatal("ReactProbe must be set by leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}
	return runReactSrc(t, req)
}

func runReactSrc(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	root := req.ModuleRoot
	reactRoot := filepath.Join(root, "react")
	if st, err := os.Stat(reactRoot); err != nil || !st.IsDir() {
		alt := filepath.Join(root, "project-api-capture-react")
		if st2, err2 := os.Stat(alt); err2 == nil && st2.IsDir() {
			reactRoot = alt
		}
	}
	resp := &Response{FileContents: map[string]string{}}

	// Primary: SessionPageApp + any detect/resolve helper modules under ui/.
	appCandidates := []string{
		filepath.Join(reactRoot, "src", "ui", "SessionPageApp.tsx"),
		filepath.Join(reactRoot, "src", "ui", "SessionPageApp.ts"),
		filepath.Join(reactRoot, "src", "ui", "SessionPageApp.jsx"),
		filepath.Join(reactRoot, "src", "apps", "session-page", "SessionPageApp.tsx"),
		filepath.Join(reactRoot, "src", "apps", "session-page", "App.tsx"),
		filepath.Join(reactRoot, "src", "apps", "session-page", "main.tsx"),
	}
	path, data, ok := firstExistingFile(appCandidates)
	if !ok {
		resp.ErrText = "SessionPageApp source not found under react/src/ui or apps/session-page"
		return resp, nil
	}
	resp.FileExists = true
	resp.FoundPaths = []string{path}
	resp.FileContents[path] = string(data)
	resp.SessionPageText = string(data)
	resp.CombinedText = string(data)

	// Prefer canonical ui/SessionPageApp when entry was main/App.
	uiApp := filepath.Join(reactRoot, "src", "ui", "SessionPageApp.tsx")
	if path != uiApp {
		if b, err := os.ReadFile(uiApp); err == nil {
			resp.FoundPaths = append(resp.FoundPaths, uiApp)
			resp.FileContents[uiApp] = string(b)
			resp.SessionPageText = resp.SessionPageText + "\n" + string(b)
			resp.CombinedText = resp.CombinedText + "\n" + string(b)
		}
	}

	// Helper modules that may host detectRuntimeBrowser / resolveInstallBrowser.
	helperNames := []string{
		"installBrowser.ts",
		"installBrowser.tsx",
		"detectInstallBrowser.ts",
		"detectInstallBrowser.tsx",
		"browserDetect.ts",
		"browserDetect.tsx",
		"resolveInstallBrowser.ts",
		"resolveInstallBrowser.tsx",
	}
	uiDir := filepath.Join(reactRoot, "src", "ui")
	for _, name := range helperNames {
		p := filepath.Join(uiDir, name)
		if _, seen := resp.FileContents[p]; seen {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		s := string(b)
		if strings.Contains(s, "detectRuntimeBrowser") ||
			strings.Contains(s, "resolveInstallBrowser") {
			resp.FoundPaths = append(resp.FoundPaths, p)
			resp.FileContents[p] = s
			resp.SessionPageText = resp.SessionPageText + "\n" + s
			resp.CombinedText = resp.CombinedText + "\n" + s
		}
	}

	// InstallGuideline for dual-path / wiring leaves.
	needGuideline := req.ReactProbe == ReactProbeWiringDualPaths ||
		req.ReactProbe == ReactProbeWiringInstallBrowser
	if needGuideline || true {
		// Always attempt load; cheap and helps CombinedText for dual-path asserts.
		for _, g := range []string{
			filepath.Join(reactRoot, "src", "ui", "InstallGuideline.tsx"),
			filepath.Join(reactRoot, "src", "ui", "InstallGuideline.ts"),
			filepath.Join(reactRoot, "src", "ui", "InstallGuideline.jsx"),
		} {
			if _, seen := resp.FileContents[g]; seen {
				continue
			}
			b, err := os.ReadFile(g)
			if err != nil {
				continue
			}
			resp.FoundPaths = append(resp.FoundPaths, g)
			resp.FileContents[g] = string(b)
			resp.GuidelineText = string(b)
			resp.CombinedText = resp.CombinedText + "\n" + string(b)
			break
		}
	}

	_ = req.ReactProbe // Assert specializes per leaf
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

// extractNamedFuncBody returns approximate source of function `name` (TS/JS).
// Matches export function name, function name, const name = (...), export const name =.
func extractNamedFuncBody(src, name string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?s)(?:export\s+)?function\s+` + regexp.QuoteMeta(name) + `\s*\([^)]*\)[^{]*\{`),
		regexp.MustCompile(`(?s)(?:export\s+)?const\s+` + regexp.QuoteMeta(name) + `\s*=\s*(?:async\s*)?\([^)]*\)\s*(?::\s*[^=]+)?\s*=>\s*\{`),
		regexp.MustCompile(`(?s)(?:export\s+)?const\s+` + regexp.QuoteMeta(name) + `\s*=\s*function\s*\([^)]*\)[^{]*\{`),
	}
	for _, re := range patterns {
		loc := re.FindStringIndex(src)
		if loc == nil {
			continue
		}
		start := loc[0]
		// Brace match from first '{' after start.
		brace := strings.Index(src[start:], "{")
		if brace < 0 {
			continue
		}
		i := start + brace
		depth := 0
		for j := i; j < len(src); j++ {
			switch src[j] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return src[start : j+1]
				}
			}
		}
	}
	return ""
}

// hasDetectRuntimeBrowser reports the pure helper name is defined.
func hasDetectRuntimeBrowser(src string) bool {
	return strings.Contains(src, "detectRuntimeBrowser")
}

// hasResolveInstallBrowser reports the priority helper name is defined.
func hasResolveInstallBrowser(src string) bool {
	return strings.Contains(src, "resolveInstallBrowser")
}

// hasFirefoxUAToken reports Firefox user-agent detection token in source.
func hasFirefoxUAToken(src string) bool {
	// Common shapes: "Firefox/", /Firefox\//, includes("Firefox"), /Firefox/i
	if strings.Contains(src, "Firefox/") {
		return true
	}
	if strings.Contains(src, `Firefox\/`) {
		return true
	}
	if strings.Contains(src, "/Firefox/") {
		return true
	}
	low := strings.ToLower(src)
	return strings.Contains(low, "firefox") &&
		(strings.Contains(src, "userAgent") || strings.Contains(src, "UserAgent") || strings.Contains(src, "navigator"))
}

// hasUserAgentRead reports navigator / userAgent access for runtime detect.
func hasUserAgentRead(src string) bool {
	return strings.Contains(src, "userAgent") ||
		strings.Contains(src, "UserAgent") ||
		strings.Contains(src, "navigator.userAgent")
}

// hasContentScriptExtMarker reports __BROWSER_AGENT_EXT__ usage.
func hasContentScriptExtMarker(src string) bool {
	return strings.Contains(src, "__BROWSER_AGENT_EXT__")
}

// indexOfFirst returns the earliest index of any needle, or -1.
func indexOfFirst(src string, needles ...string) int {
	best := -1
	for _, n := range needles {
		i := strings.Index(src, n)
		if i < 0 {
			continue
		}
		if best < 0 || i < best {
			best = i
		}
	}
	return best
}
```
