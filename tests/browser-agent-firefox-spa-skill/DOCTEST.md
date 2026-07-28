# browser-agent Firefox SPA install panel + skill docs (Phase 4 D)

Classic TDD for **Phase 4** of Firefox support: session-page **install UX**
and **skill docs** when the operator browser is Firefox.

| Surface | What is under test |
|---------|-------------------|
| React `InstallGuideline` | Firefox branch: `about:debugging`, temporary add-on, `browser-agent-firefox` |
| React `SessionPageApp` | Wires Firefox install path (boot / browsers / prop) into install panel |
| Go inject / fallback | Server HTML install markers for Firefox vs preserved Chrome path |
| Skill docs | `install-firefox-extension`, `about:debugging`, `session new --browser firefox` |

**No real Firefox. No real Chrome. No npm. No network.** Source / FS markers only.

**Classic TDD — RED** for Firefox leaves until implementer lands SPA + inject +
skill markers. Chrome-preservation leaves should stay **GREEN** (no Chrome
default path change).

### Out of scope

- Changing Chrome default install path when browser is Chrome
- New product ports
- Extension background / CDP / session-new CLI (earlier phases)

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** creates a session with **Firefox** (`session new --browser firefox`)
or the SPA/boot indicates browser **firefox**. The **session page** must show
Firefox install steps instead of the primary Chrome **chrome://extensions** /
Load unpacked path.

```text
# Firefox install mental model
Operator
  -> session new --browser firefox
  -> session page (/go?session=…)
  -> Install panel (React InstallGuideline and/or Go inject/fallback)
       1. Open about:debugging#/runtime/this-firefox (or about:debugging)
       2. Load Temporary Add-on…
       3. Select extension folder under …/browser-agent-firefox/{version}/
  -> install-firefox-extension CLI extracts path + prints same steps

# Chrome default remains
When browser is Chrome (default):
  Install panel still shows chrome://extensions + Load unpacked
  (do not replace Chrome path with Firefox-only copy)
```

**React**

```text
InstallGuideline
  when browser/product/boot indicates firefox
    -> about:debugging + temporary add-on + browser-agent-firefox path
  when chrome/default
    -> chrome://extensions + Load unpacked (preserved)

SessionPageApp
  reads boot / session snap browsers / prop
  -> passes firefox signal into InstallGuideline or renders firefox steps
```

**Go inject** (`injectSessionBoot` / `writeFallbackSessionHTML` / helpers):

```text
when session/boot browser is firefox
  install marker or panel mentions about:debugging, temporary add-on,
  browser-agent-firefox (and install-firefox-extension is welcome)
when chrome/default
  keep chrome://extensions + Load unpacked markers
```

**Skill** (`browseragent/SKILL.md` and/or `cmd/browser-agent/SKILL.md`):

```text
Documents:
  - browser-agent install-firefox-extension
  - about:debugging / Load Temporary Add-on flow
  - browser-agent session new --browser firefox
```

**Test Client** reads module sources via **ModuleRoot** =
`filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))` — never launches a
browser.

## Decision Tree

```
browser-agent-firefox-spa-skill
├── react-src/                                   [InstallGuideline + SessionPageApp FS]
│   ├── install-guideline-firefox/                 R1 firefox steps in InstallGuideline
│   ├── session-page-app-firefox/                  R2 SessionPageApp wires firefox install
│   └── chrome-guideline-preserved/                R3 chrome://extensions still primary (GREEN)
├── go-src/                                      [server inject / fallback source]
│   ├── inject-firefox-install/                    G1 firefox install markers in inject/fallback
│   └── chrome-install-preserved/                  G2 chrome://extensions still present (GREEN)
└── skill-md/                                    [SKILL.md FS]
    ├── install-firefox-extension/                 S1 install-firefox-extension documented
    ├── about-debugging-temporary-addon/           S2 about:debugging + temporary add-on
    └── session-new-browser-firefox/               S3 session new --browser firefox
```

### Parameter significance (high → low)

1. **Surface / Mode** — react-src vs go-src vs skill-md (different files / Run branches).
2. **Browser path** — Firefox install UX vs Chrome preservation (MECE operator browser).
3. **Within skill-md** — which Firefox doc marker (install CLI vs about:debugging vs session new flag).

## Test Index

| Leaf | Scenario |
|------|----------|
| `react-src/install-guideline-firefox` | (R1) `InstallGuideline` source includes Firefox install markers: `about:debugging`, temporary add-on / Load Temporary Add-on, `browser-agent-firefox` (when browser is firefox or dual-path) |
| `react-src/session-page-app-firefox` | (R2) `SessionPageApp` wires Firefox install: references `about:debugging` and/or `browser-agent-firefox` and/or browser=`firefox` / browsers / install-firefox path into install UI |
| `react-src/chrome-guideline-preserved` | (R3) `InstallGuideline` still documents `chrome://extensions` (Chrome default path not removed) |
| `go-src/inject-firefox-install` | (G1) `injectSessionBoot` / fallback / install helper sources mention `about:debugging` + temporary add-on + `browser-agent-firefox` for Firefox install UX |
| `go-src/chrome-install-preserved` | (G2) Server inject/fallback still contains `chrome://extensions` for Chrome default path |
| `skill-md/install-firefox-extension` | (S1) SKILL.md documents `install-firefox-extension` |
| `skill-md/about-debugging-temporary-addon` | (S2) SKILL.md documents `about:debugging` and temporary add-on / Load Temporary Add-on |
| `skill-md/session-new-browser-firefox` | (S3) SKILL.md documents `session new` with `--browser firefox` (or equivalent firefox session-new recipe) |

**Leaf count: 8**

## How to Run

```sh
# module root
doctest vet ./tests/browser-agent-firefox-spa-skill
doctest test ./tests/browser-agent-firefox-spa-skill
# expect RED on Firefox leaves until implementer lands feature;
# chrome-*-preserved leaves should be GREEN already

# sealed prior phases (must stay GREEN after implement)
doctest test ./tests/browser-agent-firefox-connect/...
doctest test ./tests/browser-agent-firefox-jobs/...
doctest test ./tests/browser-agent-firefox-session-new/...
doctest test ./tests/browser-agent-cli-react/spa-embed/...
doctest test ./tests/browser-agent-vite-skill/...
```

Module: `github.com/xhd2015/browser-agent`.  
Under test: `react/src/ui/InstallGuideline.tsx`, `react/src/ui/SessionPageApp.tsx`,
`browseragent/server.go` (inject/fallback), `browseragent/SKILL.md` and/or
`cmd/browser-agent/SKILL.md`.

### Implementer contract (authoritative for GREEN)

#### React — `react/src/ui/InstallGuideline.tsx`

1. When browser/session/boot indicates **firefox**, install panel must show:
   - **`about:debugging`** (prefer `about:debugging#/runtime/this-firefox`)
   - **Load Temporary Add-on** / temporary add-on wording
   - Path segment **`browser-agent-firefox`** (canonical extract dir)
2. When browser is **chrome**/default, keep **`chrome://extensions`** + Load
   unpacked (Chrome default path must remain).
3. Acceptable shapes: `browser?: "chrome" | "firefox"` prop, product field,
   or dual-path copy keyed on browser.

#### React — `react/src/ui/SessionPageApp.tsx`

4. Wire Firefox into install UX when:
   - boot / `window.__BROWSER_AGENT` / session snap has browser firefox, or
   - `browsers` includes firefox, or
   - product/extension marker indicates firefox
5. Pass firefox signal into `InstallGuideline` **or** render equivalent Firefox
   install steps (must not leave Chrome-only copy when session is Firefox).

#### Go — `browseragent/server.go` (and helpers)

6. `injectSessionBoot` and/or `writeFallbackSessionHTML` (or install-panel helper
   they call) must, for Firefox sessions/boot:
   - include **`about:debugging`**
   - mention temporary add-on / **Load Temporary Add-on**
   - mention **`browser-agent-firefox`** path (and/or `install-firefox-extension`)
7. Chrome default inject/fallback must **still** contain **`chrome://extensions`**.
8. Optional: boot JSON may include `"browser":"firefox"` when session is Firefox
   (`FormatSessionBootJSON` or inject attrs) so the SPA can detect without WS.

#### Skill — `browseragent/SKILL.md` (+ `cmd/browser-agent/SKILL.md` if mirrored)

9. Document **`browser-agent install-firefox-extension`**.
10. Document **`about:debugging`** + temporary add-on / Load Temporary Add-on.
11. Document **`browser-agent session new --browser firefox`** (flag form preferred).

**Not required:** changing Chrome-only install copy for Chrome sessions; new ports.

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
	ModeReactSrc = "react-src"
	ModeGoSrc    = "go-src"
	ModeSkillMD  = "skill-md"
)

// ReactProbe for ModeReactSrc.
const (
	ReactProbeInstallGuidelineFirefox = "install-guideline-firefox"
	ReactProbeSessionPageAppFirefox   = "session-page-app-firefox"
	ReactProbeChromeGuidelinePreserved = "chrome-guideline-preserved"
)

// GoSrcProbe for ModeGoSrc.
const (
	GoSrcInjectFirefoxInstall   = "inject-firefox-install"
	GoSrcChromeInstallPreserved = "chrome-install-preserved"
)

// SkillMDProbe for ModeSkillMD.
const (
	SkillMDInstallFirefoxExtension     = "install-firefox-extension"
	SkillMDAboutDebuggingTemporaryAddon = "about-debugging-temporary-addon"
	SkillMDSessionNewBrowserFirefox    = "session-new-browser-firefox"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	// ModuleRoot is module directory (filesystem leaves).
	// Root Setup sets from d.DOCTEST_ROOT/../..
	ModuleRoot string

	// --- react-src ---
	ReactProbe string

	// --- go-src ---
	GoSrcProbe string

	// --- skill-md ---
	SkillMDProbe string
}

// Response holds probe outcomes for all modes.
type Response struct {
	// shared FS probe
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	ErrText      string

	// skill-md convenience
	SkillFileExists bool
	SkillPath       string
	SkillText       string
	SkillPathsTried []string
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
	case ModeReactSrc:
		return runReactSrc(t, req)
	case ModeGoSrc:
		return runGoSrc(t, req)
	case ModeSkillMD:
		return runSkillMD(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- React source ---

func runReactSrc(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ReactProbe == "" {
		t.Fatal("ReactProbe must be set by leaf Setup")
	}
	root := req.ModuleRoot
	reactRoot := filepath.Join(root, "react")
	if st, err := os.Stat(reactRoot); err != nil || !st.IsDir() {
		alt := filepath.Join(root, "project-api-capture-react")
		if st2, err2 := os.Stat(alt); err2 == nil && st2.IsDir() {
			reactRoot = alt
		}
	}
	resp := &Response{FileContents: map[string]string{}}

	switch req.ReactProbe {
	case ReactProbeInstallGuidelineFirefox, ReactProbeChromeGuidelinePreserved:
		candidates := []string{
			filepath.Join(reactRoot, "src", "ui", "InstallGuideline.tsx"),
			filepath.Join(reactRoot, "src", "ui", "InstallGuideline.ts"),
			filepath.Join(reactRoot, "src", "ui", "InstallGuideline.jsx"),
			filepath.Join(reactRoot, "src", "components", "InstallGuideline.tsx"),
			filepath.Join(reactRoot, "src", "components", "InstallGuideline.jsx"),
		}
		path, data, ok := firstExistingFile(candidates)
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "InstallGuideline source not found under react/src/ui or components"
		}
		return resp, nil

	case ReactProbeSessionPageAppFirefox:
		candidates := []string{
			filepath.Join(reactRoot, "src", "ui", "SessionPageApp.tsx"),
			filepath.Join(reactRoot, "src", "ui", "SessionPageApp.ts"),
			filepath.Join(reactRoot, "src", "ui", "SessionPageApp.jsx"),
			filepath.Join(reactRoot, "src", "apps", "session-page", "SessionPageApp.tsx"),
			filepath.Join(reactRoot, "src", "apps", "session-page", "App.tsx"),
			filepath.Join(reactRoot, "src", "apps", "session-page", "main.tsx"),
		}
		path, data, ok := firstExistingFile(candidates)
		if ok {
			resp.FileExists = true
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
			// Augment with InstallGuideline so Assert can see prop wiring + steps.
			for _, ui := range []string{
				filepath.Join(reactRoot, "src", "ui", "InstallGuideline.tsx"),
				filepath.Join(reactRoot, "src", "ui", "InstallGuideline.ts"),
				filepath.Join(reactRoot, "src", "ui", "InstallGuideline.jsx"),
			} {
				if ui == path {
					continue
				}
				if b, err := os.ReadFile(ui); err == nil {
					resp.FoundPaths = append(resp.FoundPaths, ui)
					resp.FileContents[ui] = string(b)
					resp.CombinedText = resp.CombinedText + "\n" + string(b)
					break
				}
			}
			// Also pull ui/SessionPageApp if main/App was chosen first.
			uiApp := filepath.Join(reactRoot, "src", "ui", "SessionPageApp.tsx")
			if path != uiApp {
				if b, err := os.ReadFile(uiApp); err == nil {
					resp.FoundPaths = append(resp.FoundPaths, uiApp)
					resp.FileContents[uiApp] = string(b)
					resp.CombinedText = resp.CombinedText + "\n" + string(b)
				}
			}
		} else {
			resp.ErrText = "SessionPageApp source not found under react/src/ui or apps/session-page"
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown ReactProbe %q", req.ReactProbe)
	}
}

// --- Go source: inject / fallback install panels ---

func runGoSrc(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.GoSrcProbe == "" {
		t.Fatal("GoSrcProbe must be set by leaf Setup")
	}
	root := req.ModuleRoot
	resp := &Response{FileContents: map[string]string{}}

	// Prefer server.go; also fold in any sibling that mentions install / inject markers.
	candidates := []string{
		filepath.Join(root, "browseragent", "server.go"),
		filepath.Join(root, "browseragent", "session_page.go"),
		filepath.Join(root, "browseragent", "session_page_fixture.go"),
		filepath.Join(root, "browseragent", "extension_firefox.go"),
		filepath.Join(root, "browseragent", "embed.go"),
	}
	var parts []string
	var paths []string
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		paths = append(paths, p)
		resp.FileContents[p] = string(data)
		parts = append(parts, string(data))
	}
	// Walk browseragent/*.go for injectSessionBoot / writeFallbackSessionHTML / firefox install HTML.
	dir := filepath.Join(root, "browseragent")
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		p := filepath.Join(dir, e.Name())
		if _, seen := resp.FileContents[p]; seen {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		s := string(b)
		if strings.Contains(s, "injectSessionBoot") ||
			strings.Contains(s, "writeFallbackSessionHTML") ||
			strings.Contains(s, "about:debugging") ||
			strings.Contains(s, "browser-agent-install") ||
			strings.Contains(s, "chrome://extensions") {
			paths = append(paths, p)
			resp.FileContents[p] = s
			parts = append(parts, s)
		}
	}

	resp.FoundPaths = paths
	resp.FileExists = len(paths) > 0
	resp.CombinedText = strings.Join(parts, "\n")
	if !resp.FileExists {
		resp.ErrText = "no browseragent Go sources for inject/fallback install found"
	}
	_ = req.GoSrcProbe // Assert specializes
	return resp, nil
}

// --- SKILL.md ---

func runSkillMD(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SkillMDProbe == "" {
		t.Fatal("SkillMDProbe must be set by leaf Setup")
	}
	root := req.ModuleRoot
	candidates := []string{
		filepath.Join(root, "browseragent", "SKILL.md"),
		filepath.Join(root, "cmd", "browser-agent", "SKILL.md"),
	}
	resp := &Response{SkillPathsTried: candidates, FileContents: map[string]string{}}
	// Prefer combining both when present so mirrored docs both count.
	var parts []string
	var paths []string
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		paths = append(paths, p)
		resp.FileContents[p] = string(data)
		parts = append(parts, string(data))
	}
	resp.SkillFileExists = len(paths) > 0
	resp.FileExists = resp.SkillFileExists
	resp.FoundPaths = paths
	if resp.SkillFileExists {
		resp.SkillPath = paths[0]
		resp.SkillText = strings.Join(parts, "\n")
		resp.CombinedText = resp.SkillText
	} else {
		resp.ErrText = "SKILL.md not found under browseragent/ or cmd/browser-agent/"
	}
	_ = req.SkillMDProbe
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
```
