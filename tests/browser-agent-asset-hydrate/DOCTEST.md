# browser-agent asset hydrate — P1..P5 + P7 (release names + docs)

Doctest tree for asset hydrate on package
`github.com/xhd2015/browser-agent/browseragent`:

| Phase | Surface |
|-------|---------|
| **P1** | `EmbedCompleteFS` / `ResolveSessionPageIndexFS` |
| **P2** | Asset cache root/write/open/complete |
| **P3** | `EnsureAsset` + httptest |
| **P4** | `ResolveSessionPage` / `ResolveExtensionDir` (embed then ensure) |
| **P5** | CLI `assets ensure|status|help` via `HandleCLI` |
| **P7** | Release archive name helper + operator docs (hydrate/cache/ensure) |

**No real GitHub network in tests.** **No production code in this tree.**

## Mode

- **P1–P5 leaves**: keep **GREEN-compatible** — do not weaken.
- **P7 leaves**: **Classic TDD** — release name helper + hydrate docs not
  present yet. Expect **RED** until implementer lands them.

## Version

0.0.2

# DSN (Domain Specific Notion)

Prior phases: completeness, cache, download, implicit resolve, assets CLI.

**P7 Release assets + docs**:

```text
// Archive basename matches EnsureAsset URL shape (no path prefix):
//   {product}_v{version}_{kind}.tar.gz
// e.g. browser-agent_v0.2.0_session-page.tar.gz
//      browser-agent_v0.2.0_extension.tar.gz
//      browser-trace_v0.2.0_extension.tar.gz  (trace; optional but recommended)

AssetReleaseNames(version string) []string
// version "0.2.0" or "v0.2.0" → names with normalized v-prefix
```

**Operator docs** (markdown under `docs/`, `README.md`, and/or `browseragent/SKILL.md`):

- Incomplete embed / `go install` → hydrate from GitHub into
  `~/.cache/browser-agent` (or `asset-cache` path)
- Fat release offline (embed complete)
- CLI `assets ensure`
- Env `BROWSER_AGENT_ASSET_BASE_URL` and/or HTTPS proxy env notes

**Test Client**: package pure helpers + ModuleRoot filesystem probes. No Chrome.

## Decision Tree

```
browser-agent-asset-hydrate
├── completeness/ …                              [P1]
├── resolve/ …                                   [P1]
├── cache/ …                                     [P2]
├── download/ …                                  [P3]
├── implicit/ …                                  [P4]
├── cli/ …                                       [P5]
├── release/                                     [P7 archive names]
│   └── asset-names/                               names for v0.2.0 agent kinds
└── docs/                                        [P7 operator docs]
    └── assets-hydrate/                            cache + ensure + go install
```

## Test Index

| Leaf | Scenario |
|------|----------|
| P1–P5 leaves | unchanged (20) |
| `release/asset-names` | (P7) `AssetReleaseNames("v0.2.0")` includes browser-agent session-page + extension archive basenames |
| `docs/assets-hydrate` | (P7) docs/SKILL/README mention cache path, go install/incomplete embed, assets ensure |

**Leaf count: 22** (20 prior + 2 P7)

## How to Run

```sh
doctest vet ./tests/browser-agent-asset-hydrate
doctest test ./tests/browser-agent-asset-hydrate
# prior phases may be GREEN; P7 expect RED until release names + docs land
```

### Implementer contract

#### P1–P5 (keep GREEN)

```text
EmbedCompleteFS, ResolveSessionPageIndexFS
AssetCacheRoot, AssetCacheDir, WriteAssetCache, OpenAssetCache, CacheComplete
AssetDownloadConfig, EnsureAsset
ResolveSessionPage, ResolveExtensionDir
HandleCLI assets ensure|status|help
```

#### P7 release names + docs (RED until landed)

```text
// AssetReleaseNames returns archive basenames for a release version matching
// EnsureAsset download names: {product}_v{version}_{kind}.tar.gz
// Must include at least:
//   browser-agent_v0.2.0_session-page.tar.gz
//   browser-agent_v0.2.0_extension.tar.gz
// when version is "v0.2.0" or "0.2.0" (normalize to leading v).
// Recommended also: browser-trace_v0.2.0_extension.tar.gz
func AssetReleaseNames(version string) []string
// Alias OK if coherent: ReleaseAssetArchiveNames, ListAssetReleaseArchives, etc.
// — tests call browseragent.AssetReleaseNames.

// Docs (one or more files under module):
//   docs/**/*.md, README.md, browseragent/SKILL.md, SKILL.md
// Must mention (combined text across found files):
//   - cache path: ~/.cache/browser-agent and/or asset-cache
//   - go install and/or incomplete embed fallback
//   - assets ensure
// Optional: BROWSER_AGENT_ASSET_BASE_URL, HTTPS_PROXY
```

**Non-goals for this leaf set**: implementing full GH Actions release upload (may be follow-up); browser-trace full CLI (P6).

```go
import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level API surface under test.
const (
	ModeCompleteness = "completeness"
	ModeResolve      = "resolve"
	ModeCache        = "cache"
	ModeDownload     = "download"
	ModeImplicit     = "implicit"
	ModeCLI          = "cli"
	ModeRelease      = "release"
	ModeDocs         = "docs"
)

// CacheOp — P2
const (
	CacheOpRootRespectsXDG  = "root-respects-xdg"
	CacheOpWriteThenOpenHit = "write-then-open-hit"
	CacheOpMissIncomplete   = "miss-incomplete"
	CacheOpProductIsolation = "product-isolation"
)

// DownloadOp — P3
const (
	DownloadOpEnsureFillsCache      = "ensure-fills-cache"
	DownloadOpSecondEnsureNoRefetch = "second-ensure-no-refetch"
	DownloadOpOfflineOr404          = "offline-or-404-errors"
	DownloadOpVersionPinPath        = "version-pin-path"
)

// ImplicitOp — P4
const (
	ImplicitOpCompleteEmbedNoDownload       = "complete-embed-no-download"
	ImplicitOpIncompleteDownloadsThenServes = "incomplete-downloads-then-serves"
	ImplicitOpExtensionIncompleteEnsures    = "extension-incomplete-ensures"
)

// CLIOp — P5
const (
	CLIOpAssetsHelp            = "assets-help"
	CLIOpAssetsStatusEmbed     = "assets-status-embed"
	CLIOpAssetsEnsureDownloads = "assets-ensure-downloads"
)

// ReleaseOp — P7
const (
	ReleaseOpAssetNames = "asset-names"
)

// DocsOp — P7
const (
	DocsOpAssetsHydrate = "assets-hydrate"
)

// Expected archive basenames for v0.2.0 (EnsureAsset shape).
const (
	ArchiveAgentSessionPage = "browser-agent_v0.2.0_session-page.tar.gz"
	ArchiveAgentExtension   = "browser-agent_v0.2.0_extension.tar.gz"
	ArchiveTraceExtension   = "browser-trace_v0.2.0_extension.tar.gz"
)

const (
	KindSessionPage = "session-page"
	KindExtension   = "extension"
)

const (
	ProductBrowserAgent = "browser-agent"
	ProductBrowserTrace = "browser-trace"
)

const CacheVersion = "v0.2.0"

const (
	FixtureSessionPageComplete = "session-page-complete"
	FixtureSessionPageHTMLOnly = "session-page-html-only"
	FixtureExtensionComplete   = "extension-complete"
	FixtureEmpty               = "empty"
)

const (
	SessionPageRootMarker = "data-browser-agent-root"
	ResolveSourceEmbed    = "embed"
	ResolveSourceCache    = "cache"
)

// Env keys for CLI / download.
const (
	EnvXDGCacheHome           = "XDG_CACHE_HOME"
	EnvBrowserAgentAssetBase  = "BROWSER_AGENT_ASSET_BASE_URL"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode       string
	ModuleRoot string

	// completeness
	AssetKind      string
	FixtureName    string
	ExpectComplete bool

	// resolve (P1)
	ResolveFixtureName string
	ExpectResolveOK    bool

	// cache (P2)
	CacheOp           string
	CacheProduct      string
	CacheVersion      string
	CacheKind         string
	CacheWriteFixture string
	XDGCacheHome      string
	IsolateHome       bool

	// download (P3)
	DownloadOp          string
	DownloadProduct     string
	DownloadVersion     string
	DownloadKind        string
	DownloadFixture     string
	DownloadServe404    bool
	DownloadCloseServer bool
	DownloadCallTwice   bool

	// implicit (P4)
	ImplicitOp           string
	ImplicitEmbedFixture string
	ImplicitVersion      string
	ImplicitBaseDir      string
	ImplicitStartServer  bool
	ImplicitServeFixture string

	// cli (P5)
	CLIOp   string
	CLIArgs []string
	CLIEnv  map[string]string
	// CLIServeBothArchives when true, httptest serves session-page + extension tar.gz
	// based on request path; sets BROWSER_AGENT_ASSET_BASE_URL in CLIEnv.
	CLIServeBothArchives bool

	// release (P7)
	ReleaseOp      string
	ReleaseVersion string

	// docs (P7)
	DocsOp string
}

// Response holds outcomes for all modes.
type Response struct {
	// completeness
	Complete    bool
	CompleteSP  bool
	CompleteExt bool
	BothKinds   bool

	// resolve / implicit session-page
	HTML       string
	Source     string
	ResolveErr error

	// cache (P2)
	CacheRoot          string
	CacheDir           string
	WriteDir           string
	WriteErr           error
	OpenOK             bool
	OpenDir            string
	OpenErr            error
	CacheComplete      bool
	OtherOpenOK        bool
	OtherCacheComplete bool
	OpenOK2            bool
	OpenDir2           string

	// download (P3)
	EnsureDir          string
	EnsureErr          error
	EnsureDir2         string
	EnsureErr2         error
	GETCount           int64
	LastRequestURL     string
	LastRequestPath    string
	CacheCompleteAfter bool

	// implicit extension
	InstallPath string
	InstallVer  string
	InstallErr  error
	ExtComplete bool

	// cli (P5)
	CLIStdout string
	CLIStderr string
	CLIErr    error
	// After ensure CLI: both kinds complete?
	CacheCompleteSP  bool
	CacheCompleteExt bool

	// release (P7)
	ReleaseNames []string

	// docs (P7)
	DocsPaths        []string
	DocsCombinedText string
	DocsFound        bool

	// shared
	FSRoot   string
	ErrText  string
	ExitCode int
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
	case ModeCompleteness:
		return runCompleteness(t, req)
	case ModeResolve:
		return runResolve(t, req)
	case ModeCache:
		return runCache(t, req)
	case ModeDownload:
		return runDownload(t, req)
	case ModeImplicit:
		return runImplicit(t, req)
	case ModeCLI:
		return runCLI(t, req)
	case ModeRelease:
		return runRelease(t, req)
	case ModeDocs:
		return runDocs(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runCompleteness(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	fsys, root, err := openFixtureFS(t, req.FixtureName)
	if err != nil {
		return nil, err
	}
	resp := &Response{FSRoot: root, ExitCode: 0}
	if req.AssetKind == "" {
		resp.BothKinds = true
		resp.CompleteSP = browseragent.EmbedCompleteFS(fsys, KindSessionPage)
		resp.CompleteExt = browseragent.EmbedCompleteFS(fsys, KindExtension)
		return resp, nil
	}
	resp.Complete = browseragent.EmbedCompleteFS(fsys, req.AssetKind)
	return resp, nil
}

func runResolve(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	name := req.ResolveFixtureName
	if name == "" {
		name = req.FixtureName
	}
	fsys, root, err := openFixtureFS(t, name)
	if err != nil {
		return nil, err
	}
	html, source, rerr := browseragent.ResolveSessionPageIndexFS(fsys)
	return &Response{
		FSRoot:     root,
		HTML:       html,
		Source:     source,
		ResolveErr: rerr,
		ErrText:    errString(rerr),
		ExitCode:   0,
	}, nil
}

func runCache(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CacheOp == "" {
		t.Fatal("CacheOp must be set by leaf Setup")
	}
	applyCacheEnv(t, req)

	product := req.CacheProduct
	if product == "" {
		product = ProductBrowserAgent
	}
	version := req.CacheVersion
	if version == "" {
		version = CacheVersion
	}
	kind := req.CacheKind
	if kind == "" {
		kind = KindSessionPage
	}

	resp := &Response{ExitCode: 0}

	switch req.CacheOp {
	case CacheOpRootRespectsXDG:
		resp.CacheRoot = browseragent.AssetCacheRoot()
		resp.CacheDir = browseragent.AssetCacheDir(product, version, kind)
		return resp, nil

	case CacheOpWriteThenOpenHit:
		src, _, err := openFixtureFS(t, req.CacheWriteFixture)
		if err != nil {
			return nil, err
		}
		wdir, werr := browseragent.WriteAssetCache(product, version, kind, src)
		resp.WriteDir = wdir
		resp.WriteErr = werr
		resp.ErrText = errString(werr)
		if werr != nil {
			return resp, nil
		}
		fsys, odir, ok, oerr := browseragent.OpenAssetCache(product, version, kind)
		resp.OpenOK = ok
		resp.OpenDir = odir
		resp.OpenErr = oerr
		_ = fsys
		resp.CacheComplete = browseragent.CacheComplete(product, version, kind)
		_, odir2, ok2, _ := browseragent.OpenAssetCache(product, version, kind)
		resp.OpenOK2 = ok2
		resp.OpenDir2 = odir2
		return resp, nil

	case CacheOpMissIncomplete:
		resp.CacheDir = browseragent.AssetCacheDir(product, version, kind)
		resp.CacheComplete = browseragent.CacheComplete(product, version, kind)
		_, odir, ok, oerr := browseragent.OpenAssetCache(product, version, kind)
		resp.OpenOK = ok
		resp.OpenDir = odir
		resp.OpenErr = oerr
		return resp, nil

	case CacheOpProductIsolation:
		src, _, err := openFixtureFS(t, req.CacheWriteFixture)
		if err != nil {
			return nil, err
		}
		wdir, werr := browseragent.WriteAssetCache(ProductBrowserAgent, version, kind, src)
		resp.WriteDir = wdir
		resp.WriteErr = werr
		if werr != nil {
			resp.ErrText = errString(werr)
			return resp, nil
		}
		_, odir, ok, oerr := browseragent.OpenAssetCache(ProductBrowserAgent, version, kind)
		resp.OpenOK = ok
		resp.OpenDir = odir
		resp.OpenErr = oerr
		resp.CacheComplete = browseragent.CacheComplete(ProductBrowserAgent, version, kind)
		_, _, okOther, _ := browseragent.OpenAssetCache(ProductBrowserTrace, version, kind)
		resp.OtherOpenOK = okOther
		resp.OtherCacheComplete = browseragent.CacheComplete(ProductBrowserTrace, version, kind)
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown CacheOp %q", req.CacheOp)
	}
}

func runDownload(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.DownloadOp == "" {
		t.Fatal("DownloadOp must be set by leaf Setup")
	}
	applyCacheEnv(t, req)

	product := req.DownloadProduct
	if product == "" {
		product = ProductBrowserAgent
	}
	version := req.DownloadVersion
	if version == "" {
		version = CacheVersion
	}
	kind := req.DownloadKind
	if kind == "" {
		kind = KindSessionPage
	}
	fixture := req.DownloadFixture
	if fixture == "" {
		fixture = FixtureSessionPageComplete
	}

	resp := &Response{ExitCode: 0}

	var tarBytes []byte
	if !req.DownloadServe404 {
		b, err := buildFixtureTarGZ(t, fixture)
		if err != nil {
			return nil, err
		}
		tarBytes = b
	}

	var getCount atomic.Int64
	var lastURL atomic.Value
	var lastPath atomic.Value

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == "" {
			getCount.Add(1)
			lastURL.Store(r.URL.String())
			lastPath.Store(r.URL.Path)
		}
		if req.DownloadServe404 {
			http.NotFound(w, r)
			return
		}
		p := r.URL.Path
		if !strings.HasSuffix(p, ".tar.gz") && !strings.Contains(p, ".tar.gz") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(tarBytes)
	})

	srv := httptest.NewServer(mux)
	if req.DownloadCloseServer {
		srv.Close()
	} else {
		t.Cleanup(srv.Close)
	}

	baseURL := strings.TrimRight(srv.URL, "/") + "/releases/download"
	cfg := browseragent.AssetDownloadConfig{
		BaseURL:    baseURL,
		HTTPClient: srv.Client(),
	}
	if req.DownloadCloseServer {
		cfg.HTTPClient = &http.Client{Timeout: 2 * time.Second}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	dir1, err1 := browseragent.EnsureAsset(ctx, product, version, kind, cfg)
	resp.EnsureDir = dir1
	resp.EnsureErr = err1
	resp.ErrText = errString(err1)

	if req.DownloadCallTwice {
		dir2, err2 := browseragent.EnsureAsset(ctx, product, version, kind, cfg)
		resp.EnsureDir2 = dir2
		resp.EnsureErr2 = err2
	}

	resp.GETCount = getCount.Load()
	if v, ok := lastURL.Load().(string); ok {
		resp.LastRequestURL = v
	}
	if v, ok := lastPath.Load().(string); ok {
		resp.LastRequestPath = v
	}
	resp.CacheCompleteAfter = browseragent.CacheComplete(product, version, kind)
	return resp, nil
}

func runImplicit(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ImplicitOp == "" {
		t.Fatal("ImplicitOp must be set by leaf Setup")
	}
	applyCacheEnv(t, req)

	version := req.ImplicitVersion
	if version == "" {
		version = CacheVersion
	}

	embedName := req.ImplicitEmbedFixture
	if embedName == "" {
		embedName = FixtureEmpty
	}
	embedFS, embedRoot, err := openFixtureFS(t, embedName)
	if err != nil {
		return nil, err
	}

	resp := &Response{FSRoot: embedRoot, ExitCode: 0}

	var getCount atomic.Int64
	var cfg browseragent.AssetDownloadConfig
	needServer := req.ImplicitStartServer ||
		req.ImplicitOp == ImplicitOpIncompleteDownloadsThenServes ||
		req.ImplicitOp == ImplicitOpExtensionIncompleteEnsures

	if needServer {
		serveFix := req.ImplicitServeFixture
		if serveFix == "" {
			if req.ImplicitOp == ImplicitOpExtensionIncompleteEnsures {
				serveFix = FixtureExtensionComplete
			} else {
				serveFix = FixtureSessionPageComplete
			}
		}
		tarBytes, err := buildFixtureTarGZ(t, serveFix)
		if err != nil {
			return nil, err
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == "" {
				getCount.Add(1)
			}
			p := r.URL.Path
			if !strings.HasSuffix(p, ".tar.gz") && !strings.Contains(p, ".tar.gz") {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(tarBytes)
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)
		cfg = browseragent.AssetDownloadConfig{
			BaseURL:    strings.TrimRight(srv.URL, "/") + "/releases/download",
			HTTPClient: srv.Client(),
		}
	}

	switch req.ImplicitOp {
	case ImplicitOpCompleteEmbedNoDownload:
		if !needServer {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				getCount.Add(1)
				http.NotFound(w, r)
			}))
			t.Cleanup(srv.Close)
			cfg = browseragent.AssetDownloadConfig{
				BaseURL:    strings.TrimRight(srv.URL, "/") + "/releases/download",
				HTTPClient: srv.Client(),
			}
		}
		html, source, rerr := browseragent.ResolveSessionPage(embedFS, version, cfg)
		resp.HTML = html
		resp.Source = source
		resp.ResolveErr = rerr
		resp.ErrText = errString(rerr)
		resp.GETCount = getCount.Load()
		return resp, nil

	case ImplicitOpIncompleteDownloadsThenServes:
		html, source, rerr := browseragent.ResolveSessionPage(embedFS, version, cfg)
		resp.HTML = html
		resp.Source = source
		resp.ResolveErr = rerr
		resp.ErrText = errString(rerr)
		resp.GETCount = getCount.Load()
		resp.CacheCompleteAfter = browseragent.CacheComplete(ProductBrowserAgent, version, KindSessionPage)
		return resp, nil

	case ImplicitOpExtensionIncompleteEnsures:
		baseDir := req.ImplicitBaseDir
		if baseDir == "" {
			baseDir = t.TempDir()
		}
		path, ver, ierr := browseragent.ResolveExtensionDir(embedFS, baseDir, version, cfg)
		resp.InstallPath = path
		resp.InstallVer = ver
		resp.InstallErr = ierr
		resp.ErrText = errString(ierr)
		resp.GETCount = getCount.Load()
		if path != "" {
			resp.ExtComplete = browseragent.EmbedCompleteFS(os.DirFS(path), KindExtension)
		}
		resp.CacheCompleteAfter = browseragent.CacheComplete(ProductBrowserAgent, version, KindExtension)
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown ImplicitOp %q", req.ImplicitOp)
	}
}

func runCLI(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLIOp == "" {
		t.Fatal("CLIOp must be set by leaf Setup")
	}
	if len(req.CLIArgs) == 0 {
		t.Fatal("CLIArgs must be set by leaf Setup")
	}

	env := map[string]string{}
	for k, v := range req.CLIEnv {
		env[k] = v
	}
	if req.XDGCacheHome != "" {
		env[EnvXDGCacheHome] = req.XDGCacheHome
		// Also set process env so AssetCacheRoot (if it reads process env) sees isolation
		// when HandleCLI merges env into process — tests prefer env map.
		t.Setenv(EnvXDGCacheHome, req.XDGCacheHome)
	}

	if req.CLIServeBothArchives {
		spTar, err := buildFixtureTarGZ(t, FixtureSessionPageComplete)
		if err != nil {
			return nil, err
		}
		extTar, err := buildFixtureTarGZ(t, FixtureExtensionComplete)
		if err != nil {
			return nil, err
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			p := r.URL.Path
			if !strings.Contains(p, ".tar.gz") {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/gzip")
			if strings.Contains(p, "extension") {
				_, _ = w.Write(extTar)
				return
			}
			// default session-page
			_, _ = w.Write(spTar)
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)
		base := strings.TrimRight(srv.URL, "/") + "/releases/download"
		env[EnvBrowserAgentAssetBase] = base
		t.Setenv(EnvBrowserAgentAssetBase, base)
	}

	var stdout, stderr bytes.Buffer
	cliErr := browseragent.HandleCLI(req.CLIArgs, env, &stdout, &stderr)

	resp := &Response{
		CLIStdout: stdout.String(),
		CLIStderr: stderr.String(),
		CLIErr:    cliErr,
		ErrText:   errString(cliErr),
		ExitCode:  0,
	}

	// After ensure: report cache completeness for current product version.
	ver := CacheVersion
	if v := strings.TrimSpace(browseragent.ClientVersion()); v != "" {
		if !strings.HasPrefix(v, "v") {
			ver = "v" + v
		} else {
			ver = v
		}
	}
	resp.CacheCompleteSP = browseragent.CacheComplete(ProductBrowserAgent, ver, KindSessionPage)
	resp.CacheCompleteExt = browseragent.CacheComplete(ProductBrowserAgent, ver, KindExtension)
	return resp, nil
}

func runRelease(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ReleaseOp == "" {
		t.Fatal("ReleaseOp must be set by leaf Setup")
	}
	ver := req.ReleaseVersion
	if ver == "" {
		ver = CacheVersion
	}
	resp := &Response{ExitCode: 0}
	switch req.ReleaseOp {
	case ReleaseOpAssetNames:
		// Classic TDD: AssetReleaseNames may not exist yet → compile RED until landed.
		names := browseragent.AssetReleaseNames(ver)
		resp.ReleaseNames = append([]string(nil), names...)
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown ReleaseOp %q", req.ReleaseOp)
	}
}

func runDocs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.DocsOp == "" {
		t.Fatal("DocsOp must be set by leaf Setup")
	}
	root := req.ModuleRoot
	if root == "" {
		root = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}
	resp := &Response{ExitCode: 0, DocsPaths: nil}

	// Candidate operator-facing docs (first pass: known paths; then walk docs/).
	candidates := []string{
		filepath.Join(root, "docs", "assets-hydrate.md"),
		filepath.Join(root, "docs", "asset-hydrate.md"),
		filepath.Join(root, "docs", "ASSETS.md"),
		filepath.Join(root, "docs", "assets.md"),
		filepath.Join(root, "README.md"),
		filepath.Join(root, "browseragent", "SKILL.md"),
		filepath.Join(root, "SKILL.md"),
		filepath.Join(root, "browseragent", "README.md"),
	}
	var combined strings.Builder
	seen := map[string]bool{}
	addFile := func(p string) {
		if seen[p] {
			return
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return
		}
		seen[p] = true
		resp.DocsPaths = append(resp.DocsPaths, p)
		resp.DocsFound = true
		combined.Write(b)
		combined.WriteByte('\n')
	}
	for _, p := range candidates {
		addFile(p)
	}
	// Walk docs/ for any *.md mentioning assets/hydrate/cache.
	docsDir := filepath.Join(root, "docs")
	_ = filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		addFile(path)
		return nil
	})
	resp.DocsCombinedText = combined.String()
	return resp, nil
}

func applyCacheEnv(t *testing.T, req *Request) {
	t.Helper()
	if req.XDGCacheHome != "" {
		t.Setenv("XDG_CACHE_HOME", req.XDGCacheHome)
	}
	if req.IsolateHome {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		if req.XDGCacheHome == "" {
			t.Setenv("XDG_CACHE_HOME", "")
		}
	}
}

func openFixtureFS(t *testing.T, fixtureName string) (fs.FS, string, error) {
	t.Helper()
	if fixtureName == "" {
		return nil, "", fmt.Errorf("fixture name is required")
	}
	if fixtureName == FixtureEmpty {
		dir := t.TempDir()
		return os.DirFS(dir), dir, nil
	}
	root := filepath.Join(DOCTEST_ROOT, "testdata", fixtureName)
	st, err := os.Stat(root)
	if err != nil {
		return nil, "", fmt.Errorf("fixture %q: %w", fixtureName, err)
	}
	if !st.IsDir() {
		return nil, "", fmt.Errorf("fixture %q is not a directory", fixtureName)
	}
	return os.DirFS(root), root, nil
}

func buildFixtureTarGZ(t *testing.T, fixtureName string) ([]byte, error) {
	t.Helper()
	root := filepath.Join(DOCTEST_ROOT, "testdata", fixtureName)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = rel
		if d.IsDir() {
			hdr.Name = strings.TrimSuffix(hdr.Name, "/") + "/"
			return tw.WriteHeader(hdr)
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
	if err != nil {
		_ = tw.Close()
		_ = gz.Close()
		return nil, err
	}
	if err := tw.Close(); err != nil {
		_ = gz.Close()
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func incompleteErrorMessage(err error) bool {
	if err == nil {
		return false
	}
	low := strings.ToLower(err.Error())
	return strings.Contains(low, "incomplete") ||
		strings.Contains(low, "not available") ||
		strings.Contains(low, "embed incomplete") ||
		strings.Contains(low, "not complete")
}

func downloadErrorMessage(err error) bool {
	if err == nil {
		return false
	}
	low := strings.ToLower(err.Error())
	return strings.Contains(low, "404") ||
		strings.Contains(low, "not found") ||
		strings.Contains(low, "download") ||
		strings.Contains(low, "http") ||
		strings.Contains(low, "status") ||
		strings.Contains(low, "connect") ||
		strings.Contains(low, "refused") ||
		strings.Contains(low, "ensure")
}
```
