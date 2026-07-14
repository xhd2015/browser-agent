# browser-agent bundle pipeline + dual-product coexistence

Exercises the **fixture Bundle pipeline** for browser-agent embed staging and
**dual-product** non-collision (browser-agent vs browser-trace / Capture-API).

| Surface | What is under test |
|---------|-------------------|
| Bundle fixture | Package `Bundle(UseFixture)` stages extension + session-page under isolated Root |
| Dual ProductConfig | Go `ProductBrowserAgent` (43761) vs `ProductBrowserTrace` (43759) |
| Dual extension shells | On-disk `Chrome-Ext-Browser-Agent` vs `Chrome-Ext-Capture-API` ports/markers |
| Dual React products | `react/src/products/browser-agent.ts` vs `browser-trace.ts` ports |

**No npm. No real Chrome. No real agent-run.** Bundle leaves use **temp Root**
and fixture sources (no write into live repo embed when isolation options are
honored). Filesystem leaves read ModuleRoot only.

**Sealed** prior trees (do **not** modify; must stay GREEN):

```sh
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-cli-react/...
doctest test ./tests/browser-agent-serve-runtime/...
doctest test ./tests/browser-agent-cdp-jobs/...
doctest test ./tests/browser-agent-vite-skill/...
# browser-trace regressions as needed
```

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator / CI** prepares embed payloads for `browseragent` via a **Bundle
pipeline** (package API preferred; thin `script/browser-agent/bundle` optional):

```text
Bundle(BundleOptions{Root, UseFixture: true, …})
  -> stage extension into  {Root}/browseragent/embedded/extension/
  -> stage session-page into {Root}/browseragent/embedded/session-page/
  -> BundleResult{ExtensionDir, SessionPageDir, UsedFixture}
```

- **UseFixture=true**: copy mini fixtures (no npm). Idempotent: second Bundle
  succeeds with stable destination paths.
- Staged extension content targets control port **43761** (browser-agent).
- Isolation: tests pass **Root = temp dir** and absolute **Fixture*Dir**
  overrides so live module embed is not mutated.

**ProductConfig** (Go package + React TS mirror) keeps two products distinct:

| Product | Port | Page marker | Feature | Extension dir |
|---------|------|-------------|---------|---------------|
| browser-agent | **43761** | `__BROWSER_AGENT_EXT__` | `browser-agent` | `Chrome-Ext-Browser-Agent` |
| browser-trace | **43759** | `__BROWSER_TRACE_EXT__` | `browser-trace` | `Chrome-Ext-Capture-API` |

**On-disk shells** under ModuleRoot must match those ports/markers (manifest +
content scripts). **React products** under `react/src/products/*` mirror ports.

**Test Client** in this tree:

- Bundle leaves call package `browseragent.Bundle` with temp Root + fixture paths
- Product-go leaves read `ProductBrowserAgent` / `ProductBrowserTrace`
- Shell-disk / react-products leaves read ModuleRoot files only

## Decision Tree

```
browser-agent-bundle-dual
├── bundle-fixture/                              [ModeBundle UseFixture]
│   ├── extension-manifest/                        A1 ExtensionDir has manifest; version non-empty
│   ├── session-page-root/                         A2 SessionPageDir index.html root mount
│   ├── idempotent-twice/                          A3 Bundle twice → stable paths
│   └── extension-hosts-43761/                     A4 staged extension mentions 43761
├── dual-product-go/                             [ModeProductGo]
│   ├── browser-agent/                             C1 port 43761, feature browser-agent
│   ├── browser-trace/                             C2 port 43759, feature browser-trace
│   └── ports-differ/                              C3 AgentPort != TracePort
├── dual-shell-disk/                             [ModeShellDisk ModuleRoot FS]
│   ├── agent-manifest/                            D1 Chrome-Ext-Browser-Agent: 43761 + Browser Agent
│   ├── capture-manifest/                          D2 Chrome-Ext-Capture-API: 43759
│   ├── agent-content-script/                      D3 __BROWSER_AGENT_EXT__ + browser-agent
│   └── capture-content-script/                    D4 __BROWSER_TRACE_EXT__ + browser-trace
└── react-products/                              [ModeReactProducts]
    ├── browser-agent/                             E1 react product file contains 43761
    └── browser-trace/                             E2 react product file contains 43759
```

### Parameter significance (high → low)

1. **Surface / Mode** — Bundle pipeline vs Go ProductConfig vs on-disk shells vs
   React products (different `Run` branches and contracts).
2. **Within Bundle** — assertion focus (extension manifest vs session-page vs
   idempotency vs host port) / pass count.
3. **Product id** — browser-agent vs browser-trace vs ports-differ (both).
4. **Shell probe** — manifest vs content-script for each product.

## Test Index

| Leaf | Scenario |
|------|----------|
| `bundle-fixture/extension-manifest` | (A1) `Bundle(UseFixture)` → ExtensionDir has `manifest.json`; version non-empty |
| `bundle-fixture/session-page-root` | (A2) SessionPageDir has `index.html` (or session-page.html) with root mount |
| `bundle-fixture/idempotent-twice` | (A3) Bundle twice → both succeed; ExtensionDir/SessionPageDir stable |
| `bundle-fixture/extension-hosts-43761` | (A4) Staged extension text mentions **43761** |
| `dual-product-go/browser-agent` | (C1) ProductBrowserAgent: port 43761, feature browser-agent, distinct id |
| `dual-product-go/browser-trace` | (C2) ProductBrowserTrace: port 43759, feature browser-trace |
| `dual-product-go/ports-differ` | (C3) AgentPort != TracePort (43761 vs 43759) |
| `dual-shell-disk/agent-manifest` | (D1) Chrome-Ext-Browser-Agent manifest: 43761 + Browser Agent |
| `dual-shell-disk/capture-manifest` | (D2) Chrome-Ext-Capture-API manifest: 43759 |
| `dual-shell-disk/agent-content-script` | (D3) Agent contentScript: `__BROWSER_AGENT_EXT__` + browser-agent |
| `dual-shell-disk/capture-content-script` | (D4) Capture contentScript: `__BROWSER_TRACE_EXT__` + browser-trace |
| `react-products/browser-agent` | (E1) `react/src/products/browser-agent.ts` contains 43761 |
| `react-products/browser-trace` | (E2) `react/src/products/browser-trace.ts` contains 43759 |

**Leaf count: 13**

Optional requirement **B1** (`go run ./script/browser-agent/bundle --fixture`) is
**deferred** — prefer package `Bundle` for isolation; implementer may add a thin
CLI main that calls the same API without a doctest leaf here.

## How to Run

```sh
cd project-api-capture
doctest vet ./tests/browser-agent-bundle-dual
doctest test ./tests/browser-agent-bundle-dual/...

# sealed regressions
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-cli-react/...
doctest test ./tests/browser-agent-serve-runtime/...
doctest test ./tests/browser-agent-cdp-jobs/...
doctest test ./tests/browser-agent-vite-skill/...
```

Module: `github.com/xhd2015/browser-agent`.  
Package under test: `github.com/xhd2015/browser-agent/browseragent`.

### Implementer contract (authoritative for GREEN)

#### Bundle API (required for A*)

```go
// Prefer exported from package browseragent (or browseragent/bundle re-exported):

type BundleOptions struct {
	Root        string // project root (tests use t.TempDir())
	UseFixture  bool   // stage mini fixtures; no npm

	// Optional absolute fixture sources when UseFixture (tests set these).
	// Empty → discover under Root or package defaults.
	FixtureExtensionDir   string
	FixtureSessionPageDir string

	// Dest relative to Root (defaults):
	//   EmbedExtensionRel = "browseragent/embedded/extension"
	//   EmbedSessionRel   = "browseragent/embedded/session-page"
	EmbedExtensionRel string
	EmbedSessionRel   string
}

type BundleResult struct {
	ExtensionDir   string // absolute path staged for embed
	SessionPageDir string
	UsedFixture    bool
}

func Bundle(opts BundleOptions) (*BundleResult, error)
```

**Behavior (UseFixture=true):**

1. Resolve fixture sources (opts.Fixture* absolute preferred; else defaults under
   Root / known mini paths).
2. Stage into `{Root}/{EmbedExtensionRel}` and `{Root}/{EmbedSessionRel}`
   (create parents; clear/replace contents as needed).
3. Extension dir must contain `manifest.json` with non-empty `"version"`.
4. Session-page dir must contain `index.html` or `session-page.html` with a root
   mount (`id="root"` and/or `data-browser-agent-root` / `browser-agent-root`).
5. Extension staged content must mention **43761** (hosts/permissions/matches).
6. Second call with same opts is **idempotent**: nil error; same absolute
   ExtensionDir / SessionPageDir; UsedFixture remains true.
7. Must **not** require npm, Chrome, or network.

Fixture sources acceptable for CI (tests wire absolute paths from ModuleRoot):

- Extension: `browseragent/embedded/extension` or
  `tests/browser-agent-cli-react/testdata/mini-extension`
- Session-page: `browseragent/embedded/session-page`

Optional: `script/browser-agent/bundle` main that calls `Bundle` with module cwd
and `--fixture` / `--mini` flag (not covered by this tree's leaves).

#### ProductConfig (C*)

Already designed; assert:

```go
var ProductBrowserAgent ProductConfig // ControlPort 43761, ID/CLI browser-agent, Features include browser-agent
var ProductBrowserTrace ProductConfig // ControlPort 43759, ID browser-trace, Features include browser-trace
// Ports must differ.
```

#### On-disk shells (D*)

Under ModuleRoot (no build):

- `Chrome-Ext-Browser-Agent/**/manifest.json` — Browser Agent + **43761**
- `Chrome-Ext-Capture-API/**/manifest.json` — **43759** (API Capture / browser-trace)
- Agent content script (`content.js` / `contentScript.js`): `__BROWSER_AGENT_EXT__` + `browser-agent`
- Capture content script (src or public): `__BROWSER_TRACE_EXT__` + `browser-trace`

#### React products (E*)

- `react/src/products/browser-agent.ts` (`.tsx`/`.js` OK) contains `43761`
- `react/src/products/browser-trace.ts` contains `43759`

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode values — top-level surface under test.
const (
	ModeBundle        = "bundle"
	ModeProductGo     = "product-go"
	ModeShellDisk     = "shell-disk"
	ModeReactProducts = "react-products"
)

// ProductProbe under ModeProductGo.
const (
	ProductProbeAgent       = "browser-agent"
	ProductProbeTrace       = "browser-trace"
	ProductProbePortsDiffer = "ports-differ"
)

// ShellProduct + ShellProbe under ModeShellDisk.
const (
	ShellProductAgent   = "browser-agent"
	ShellProductCapture = "browser-trace" // Chrome-Ext-Capture-API

	ShellProbeManifest      = "manifest"
	ShellProbeContentScript = "content-script"
)

// ReactProductID under ModeReactProducts.
const (
	ReactProductAgent = "browser-agent"
	ReactProductTrace = "browser-trace"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// ModuleRoot is project-api-capture module directory.
	// Root Setup sets from DOCTEST_ROOT/../..
	ModuleRoot string

	// --- Bundle (ModeBundle) ---
	// BundleRoot is isolated temp project root for staging (not ModuleRoot).
	BundleRoot string
	// UseFixture always true for this tree's Bundle leaves.
	UseFixture bool
	// BundlePasses: 1 default; 2 for idempotent-twice.
	BundlePasses int
	// Absolute fixture sources (set by bundle-fixture Setup from ModuleRoot).
	FixtureExtensionDir   string
	FixtureSessionPageDir string
	// Optional embed rel overrides (empty → package defaults).
	EmbedExtensionRel string
	EmbedSessionRel   string

	// --- ProductGo ---
	// ProductProbe: browser-agent | browser-trace | ports-differ
	ProductProbe string

	// --- ShellDisk ---
	ShellProduct string // browser-agent | browser-trace
	ShellProbe   string // manifest | content-script

	// --- ReactProducts ---
	ReactProductID string // browser-agent | browser-trace
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// Bundle
	ExtensionDir          string
	SessionPageDir        string
	UsedFixture           bool
	SecondExtensionDir    string
	SecondSessionPageDir  string
	SecondUsedFixture     bool
	ManifestPath          string
	ManifestText          string
	ManifestVersion       string
	SessionIndexPath      string
	SessionIndexText      string
	ExtensionCombinedText string
	ExitCode              int
	ErrText               string

	// ProductConfig
	ProductID          string
	ProductDisplayName string
	ProductCLIName     string
	ProductControlPort int
	ProductPortStr     string
	ProductFeatures    []string
	ProductPageMarker  string
	ProductExtDirName  string
	// Ports-differ leaf
	AgentPort int
	TracePort int
	AgentID   string
	TraceID   string

	// Filesystem probes (shell-disk / react-products)
	FoundPaths   []string
	FileContents map[string]string
	FileExists   bool
	CombinedText string
	// ManifestText also used for shell manifest probe
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
	case ModeBundle:
		return runBundle(t, req)
	case ModeProductGo:
		return runProductGo(t, req)
	case ModeShellDisk:
		return runShellDisk(t, req)
	case ModeReactProducts:
		return runReactProducts(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- Bundle ---

func runBundle(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BundleRoot == "" {
		t.Fatal("BundleRoot must be set by Setup (temp isolation Root)")
	}
	if !req.UseFixture {
		// This tree only covers UseFixture path.
		req.UseFixture = true
	}
	passes := req.BundlePasses
	if passes <= 0 {
		passes = 1
	}

	opts := browseragent.BundleOptions{
		Root:                  req.BundleRoot,
		UseFixture:            true,
		FixtureExtensionDir:   req.FixtureExtensionDir,
		FixtureSessionPageDir: req.FixtureSessionPageDir,
		EmbedExtensionRel:     req.EmbedExtensionRel,
		EmbedSessionRel:       req.EmbedSessionRel,
	}

	res1, err := browseragent.Bundle(opts)
	if err != nil {
		return &Response{ExitCode: 1, ErrText: err.Error()}, err
	}
	resp := fillBundleResult(res1)
	resp.ExitCode = 0

	if passes >= 2 {
		res2, err2 := browseragent.Bundle(opts)
		if err2 != nil {
			resp.ExitCode = 1
			resp.ErrText = err2.Error()
			return resp, err2
		}
		if res2 != nil {
			resp.SecondExtensionDir = res2.ExtensionDir
			resp.SecondSessionPageDir = res2.SessionPageDir
			resp.SecondUsedFixture = res2.UsedFixture
		}
	}
	return resp, nil
}

func fillBundleResult(res *browseragent.BundleResult) *Response {
	resp := &Response{}
	if res == nil {
		resp.ErrText = "Bundle returned nil result"
		resp.ExitCode = 1
		return resp
	}
	resp.ExtensionDir = res.ExtensionDir
	resp.SessionPageDir = res.SessionPageDir
	resp.UsedFixture = res.UsedFixture

	if res.ExtensionDir != "" {
		mp := filepath.Join(res.ExtensionDir, "manifest.json")
		resp.ManifestPath = mp
		if b, err := os.ReadFile(mp); err == nil {
			resp.ManifestText = string(b)
			resp.ManifestVersion = parseManifestVersion(string(b))
		}
		resp.ExtensionCombinedText = readTreeText(res.ExtensionDir, 32)
	}
	if res.SessionPageDir != "" {
		for _, name := range []string{"index.html", "session-page.html"} {
			p := filepath.Join(res.SessionPageDir, name)
			if b, err := os.ReadFile(p); err == nil {
				resp.SessionIndexPath = p
				resp.SessionIndexText = string(b)
				break
			}
		}
	}
	return resp
}

func parseManifestVersion(text string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		// Fallback: crude scan for "version": "…"
		low := text
		key := `"version"`
		i := strings.Index(low, key)
		if i < 0 {
			return ""
		}
		rest := low[i+len(key):]
		// find first quoted string
		q1 := strings.Index(rest, `"`)
		if q1 < 0 {
			return ""
		}
		rest = rest[q1+1:]
		q2 := strings.Index(rest, `"`)
		if q2 < 0 {
			return ""
		}
		return rest[:q2]
	}
	if v, ok := m["version"].(string); ok {
		return v
	}
	return ""
}

func readTreeText(root string, maxFiles int) string {
	var b strings.Builder
	n := 0
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return err
		}
		if n >= maxFiles {
			return filepath.SkipDir
		}
		// Skip binaries/icons
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".woff", ".woff2":
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		b.WriteString(string(data))
		b.WriteByte('\n')
		n++
		return nil
	})
	return b.String()
}

// --- ProductGo ---

func runProductGo(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ProductProbe == "" {
		t.Fatal("ProductProbe must be set")
	}
	resp := &Response{}
	switch req.ProductProbe {
	case ProductProbeAgent:
		cfg := browseragent.ProductBrowserAgent
		fillProduct(resp, cfg)
	case ProductProbeTrace:
		cfg := browseragent.ProductBrowserTrace
		fillProduct(resp, cfg)
	case ProductProbePortsDiffer:
		a := browseragent.ProductBrowserAgent
		tr := browseragent.ProductBrowserTrace
		resp.AgentPort = a.ControlPort
		resp.TracePort = tr.ControlPort
		resp.AgentID = a.ID
		resp.TraceID = tr.ID
		if resp.AgentID == "" {
			resp.AgentID = a.CLIName
		}
		if resp.TraceID == "" {
			resp.TraceID = tr.CLIName
		}
		// Also expose agent snapshot for convenience
		fillProduct(resp, a)
	default:
		return nil, fmt.Errorf("unknown ProductProbe %q", req.ProductProbe)
	}
	return resp, nil
}

func fillProduct(resp *Response, cfg browseragent.ProductConfig) {
	resp.ProductID = cfg.ID
	resp.ProductDisplayName = cfg.DisplayName
	resp.ProductCLIName = cfg.CLIName
	resp.ProductControlPort = cfg.ControlPort
	if cfg.ControlPort != 0 {
		resp.ProductPortStr = fmt.Sprintf("%d", cfg.ControlPort)
	}
	resp.ProductFeatures = append([]string(nil), cfg.Features...)
	resp.ProductPageMarker = cfg.PageMarkerGlobal
	resp.ProductExtDirName = cfg.ExtensionDirName
}

// --- ShellDisk ---

func runShellDisk(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ShellProduct == "" || req.ShellProbe == "" {
		t.Fatal("ShellProduct and ShellProbe must be set")
	}
	root := req.ModuleRoot
	resp := &Response{
		FileContents: map[string]string{},
	}

	var dirName string
	switch req.ShellProduct {
	case ShellProductAgent:
		dirName = "Chrome-Ext-Browser-Agent"
	case ShellProductCapture:
		dirName = "Chrome-Ext-Capture-API"
	default:
		return nil, fmt.Errorf("unknown ShellProduct %q", req.ShellProduct)
	}
	base := filepath.Join(root, dirName)
	if st, err := os.Stat(base); err != nil || !st.IsDir() {
		resp.ErrText = fmt.Sprintf("%s missing under ModuleRoot: %v", dirName, err)
		resp.FileExists = false
		return resp, nil
	}

	switch req.ShellProbe {
	case ShellProbeManifest:
		paths := findNamedUnder(base, "manifest.json", 12)
		resp.FoundPaths = paths
		for _, p := range paths {
			if b, err := os.ReadFile(p); err == nil {
				resp.FileContents[p] = string(b)
				resp.FileExists = true
				if resp.ManifestText == "" {
					resp.ManifestText = string(b)
					resp.ManifestPath = p
				}
				resp.CombinedText += string(b) + "\n"
			}
		}
	case ShellProbeContentScript:
		// Prefer contentScript.js / content.js under public/ or src/
		candidates := []string{
			"public/contentScript.js",
			"public/content.js",
			"src/contentScript.js",
			"src/content.js",
			"contentScript.js",
			"content.js",
		}
		var paths []string
		for _, rel := range candidates {
			p := filepath.Join(base, rel)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				paths = append(paths, p)
			}
		}
		// Fallback: walk for content*.js
		if len(paths) == 0 {
			paths = findContentScripts(base, 8)
		}
		resp.FoundPaths = paths
		for _, p := range paths {
			if b, err := os.ReadFile(p); err == nil {
				resp.FileContents[p] = string(b)
				resp.FileExists = true
				resp.CombinedText += string(b) + "\n"
			}
		}
	default:
		return nil, fmt.Errorf("unknown ShellProbe %q", req.ShellProbe)
	}
	return resp, nil
}

func findNamedUnder(root, name string, max int) []string {
	var out []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == "node_modules" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Name() == name {
			out = append(out, path)
			if len(out) >= max {
				return filepath.SkipAll
			}
		}
		return nil
	})
	// Prefer public/ paths first when multiple manifests exist.
	if len(out) > 1 {
		var preferred, rest []string
		for _, p := range out {
			if strings.Contains(filepath.ToSlash(p), "/public/") {
				preferred = append(preferred, p)
			} else {
				rest = append(rest, p)
			}
		}
		out = append(preferred, rest...)
	}
	return out
}

func findContentScripts(root string, max int) []string {
	var out []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == "node_modules" || base == ".git" || base == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		n := strings.ToLower(info.Name())
		if n == "contentscript.js" || n == "content.js" || strings.HasPrefix(n, "content") && strings.HasSuffix(n, ".js") {
			out = append(out, path)
			if len(out) >= max {
				return filepath.SkipDir
			}
		}
		return nil
	})
	return out
}

// --- ReactProducts ---

func runReactProducts(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ReactProductID == "" {
		t.Fatal("ReactProductID must be set")
	}
	root := req.ModuleRoot
	resp := &Response{FileContents: map[string]string{}}

	baseName := req.ReactProductID
	// Prefer react/ then project-api-capture-react/
	searchRoots := []string{
		filepath.Join(root, "react", "src", "products"),
		filepath.Join(root, "project-api-capture-react", "src", "products"),
	}
	exts := []string{".ts", ".tsx", ".js", ".jsx"}
	var paths []string
	for _, dir := range searchRoots {
		for _, ext := range exts {
			p := filepath.Join(dir, baseName+ext)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				paths = append(paths, p)
			}
		}
	}
	resp.FoundPaths = paths
	for _, p := range paths {
		if b, err := os.ReadFile(p); err == nil {
			resp.FileContents[p] = string(b)
			resp.FileExists = true
			resp.CombinedText += string(b) + "\n"
		}
	}
	return resp, nil
}
```
