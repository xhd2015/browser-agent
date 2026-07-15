# browser-trace asset hydrate — P6

Doctest tree for the **browser-trace** product surface of asset hydrate:

| Surface | Focus |
|---------|--------|
| **Completeness** | `EmbedCompleteFS` on extension fixtures (reuse `browseragent`) |
| **Download** | `browseragent.EnsureAsset` with product key `browser-trace` |
| **CLI** | `browser-trace assets ensure\|status\|--help` via package `HandleCLI` |

**Codebase fact**: `browsertrace` only embeds **extension** (`//go:embed embedded/extension/**`).
There is **no** session-page SPA embed for this product — leaves do not cover session-page.

**No real GitHub.** **No production code here.**

## Mode

**Classic TDD**. Completeness + EnsureAsset may already GREEN (shared
`browseragent` APIs). CLI leaves expect **RED** until `browsertrace.HandleCLI`
dispatches `assets` for product `browser-trace` (extension only).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** wants browser-trace extension assets available offline after a
versioned release, without shipping incomplete embeds forever.

**Shared cache root** (same as browser-agent hydrate):

```text
{XDG_CACHE_HOME|~/.cache}/browser-agent/asset-cache/{product}/{version}/{kind}
# product = browser-trace
# kind    = extension only for this product
```

**Completeness Checker** (`browseragent.EmbedCompleteFS`) reports whether an
`fs.FS` rooted at an extension tree has non-empty `manifest.json` and
`background.js`.

**Asset Downloader** (`browseragent.EnsureAsset`) GETs:

```text
{BaseURL}/v{version}/browser-trace_v{version}_extension.tar.gz
```

extracts, writes via `WriteAssetCache`, and returns the complete cache dir for
product `browser-trace`. No network when cache already complete.

**browser-trace CLI** exposes:

```text
browser-trace assets ensure [flags]
browser-trace assets status [flags]
browser-trace assets --help | assets help
```

Invoked in tests through package API (preferred over binary shell-out):

```text
browsertrace.HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
# args after binary name, e.g. ["assets", "status"]
# when env != nil, use only that map (inject XDG_CACHE_HOME, BROWSER_AGENT_ASSET_BASE_URL)
```

- **`assets ensure`**: product `browser-trace`, current version (`ClientVersion` →
  `v{version}`), ensure **extension** only via `EnsureAsset` when cache incomplete.
  Exit success = nil error. May log to stderr. Uses `BROWSER_AGENT_ASSET_BASE_URL`
  when set as download BaseURL.
- **`assets status`**: print embed complete? cache complete? path for extension;
  **no network required**. (CLI leaves in this lean tree pin help + ensure.)
- **`assets --help` / `assets help`**: mentions ensure and status; nil error;
  trailing `\n`.

**Test Client**: package APIs + buffers + temp XDG + httptest. No Chrome.

## Decision Tree

```
browser-trace-asset-hydrate
├── completeness/                                  [EmbedCompleteFS extension]
│   ├── extension-fixture/                           complete fixture → true
│   └── empty-incomplete/                            empty FS → false
├── download/                                      [EnsureAsset product=browser-trace]
│   └── ensure-extension/                            httptest + cold XDG → complete cache
└── cli/                                           [HandleCLI assets]
    ├── assets-help/                                 assets --help → ensure+status
    └── assets-ensure/                               ensure hydrates extension only
```

### Parameter significance (high → low)

1. **Surface** — completeness vs download vs CLI (different contracts).
2. **Fixture / op** — complete fixture vs empty; ensure-extension; help vs ensure.

## Test Index

| Leaf | Scenario |
|------|----------|
| `completeness/extension-fixture` | complete extension fixture → `EmbedCompleteFS` true |
| `completeness/empty-incomplete` | empty FS → `EmbedCompleteFS(extension)` false |
| `download/ensure-extension` | httptest + cold XDG → `EnsureAsset(browser-trace, …, extension)` fills complete cache |
| `cli/assets-help` | `assets --help` mentions ensure + status; nil err; trailing `\n` |
| `cli/assets-ensure` | cold cache + BASE_URL httptest → ensure nil err; extension cache complete |

**Leaf count: 5**

## How to Run

```sh
doctest vet ./tests/browser-trace-asset-hydrate
doctest test ./tests/browser-trace-asset-hydrate
# Completeness / EnsureAsset may GREEN; CLI expect RED until HandleCLI assets lands
```

### Implementer contract

```text
// Reuse (already in browseragent — do not break):
EmbedCompleteFS, EnsureAsset, CacheComplete, AssetCacheRoot, WriteAssetCache
// product string "browser-trace", kind "extension", version normalize with leading v

// browsertrace (RED until landed):
// HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
// routes first token "assets":
//   assets ensure  → EnsureAsset for browser-trace + ClientVersion + extension only
//                    (BaseURL from env BROWSER_AGENT_ASSET_BASE_URL)
//   assets status  → print embed/cache completeness + path for extension; no network
//   assets --help | assets help | assets -h → help text with ensure and status
//
// Env (when env map non-nil on HandleCLI):
//   XDG_CACHE_HOME
//   BROWSER_AGENT_ASSET_BASE_URL  // download base e.g. http://127.0.0.1:port/releases/download
//   HOME / USERPROFILE as needed for cache root
//
// Success paths return nil error; stdout help ends with trailing \n.
// Wire cmd/browser-trace main to HandleCLI for assets (and existing cmds as needed).
```

**Non-goals**: browser-agent CLI changes, session-page hydrate for browser-trace,
release scripts (P7).

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
	"github.com/xhd2015/browser-agent/browsertrace"
)

// Mode — top-level API surface under test.
const (
	ModeCompleteness = "completeness"
	ModeDownload     = "download"
	ModeCLI          = "cli"
)

// DownloadOp
const (
	DownloadOpEnsureExtension = "ensure-extension"
)

// CLIOp
const (
	CLIOpAssetsHelp   = "assets-help"
	CLIOpAssetsEnsure = "assets-ensure"
)

const (
	KindExtension = "extension"
)

const (
	ProductBrowserTrace = "browser-trace"
)

const CacheVersion = "v0.2.0"

const (
	FixtureExtensionComplete = "extension-complete"
	FixtureEmpty             = "empty"
)

// Env keys for CLI / download.
const (
	EnvXDGCacheHome          = "XDG_CACHE_HOME"
	EnvBrowserAgentAssetBase = "BROWSER_AGENT_ASSET_BASE_URL"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode       string
	ModuleRoot string

	// completeness
	AssetKind      string
	FixtureName    string
	ExpectComplete bool

	// download
	DownloadOp      string
	DownloadProduct string
	DownloadVersion string
	DownloadKind    string
	DownloadFixture string
	XDGCacheHome    string

	// cli
	CLIOp                string
	CLIArgs              []string
	CLIEnv               map[string]string
	CLIServeExtensionTar bool
}

// Response holds outcomes for all modes.
type Response struct {
	// completeness
	Complete bool

	// download
	EnsureDir          string
	EnsureErr          error
	GETCount           int64
	LastRequestPath    string
	CacheCompleteAfter bool

	// cli
	CLIStdout        string
	CLIStderr        string
	CLIErr           error
	CacheCompleteExt bool

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
	case ModeDownload:
		return runDownload(t, req)
	case ModeCLI:
		return runCLI(t, req)
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
	kind := req.AssetKind
	if kind == "" {
		kind = KindExtension
	}
	return &Response{
		FSRoot:   root,
		Complete: browseragent.EmbedCompleteFS(fsys, kind),
		ExitCode: 0,
	}, nil
}

func runDownload(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.DownloadOp == "" {
		t.Fatal("DownloadOp must be set by leaf Setup")
	}
	if req.XDGCacheHome != "" {
		t.Setenv(EnvXDGCacheHome, req.XDGCacheHome)
	}

	product := req.DownloadProduct
	if product == "" {
		product = ProductBrowserTrace
	}
	version := req.DownloadVersion
	if version == "" {
		version = CacheVersion
	}
	kind := req.DownloadKind
	if kind == "" {
		kind = KindExtension
	}
	fixture := req.DownloadFixture
	if fixture == "" {
		fixture = FixtureExtensionComplete
	}

	tarBytes, err := buildFixtureTarGZ(t, fixture)
	if err != nil {
		return nil, err
	}

	var getCount atomic.Int64
	var lastPath atomic.Value

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == "" {
			getCount.Add(1)
			lastPath.Store(r.URL.Path)
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

	baseURL := strings.TrimRight(srv.URL, "/") + "/releases/download"
	cfg := browseragent.AssetDownloadConfig{
		BaseURL:    baseURL,
		HTTPClient: srv.Client(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	dir, err1 := browseragent.EnsureAsset(ctx, product, version, kind, cfg)
	resp := &Response{
		EnsureDir:          dir,
		EnsureErr:          err1,
		ErrText:            errString(err1),
		GETCount:           getCount.Load(),
		CacheCompleteAfter: browseragent.CacheComplete(product, version, kind),
		ExitCode:           0,
	}
	if v, ok := lastPath.Load().(string); ok {
		resp.LastRequestPath = v
	}
	return resp, nil
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
		t.Setenv(EnvXDGCacheHome, req.XDGCacheHome)
	}

	if req.CLIServeExtensionTar {
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
			// Only extension archives for browser-trace product.
			if !strings.Contains(p, "extension") {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(extTar)
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)
		base := strings.TrimRight(srv.URL, "/") + "/releases/download"
		env[EnvBrowserAgentAssetBase] = base
		t.Setenv(EnvBrowserAgentAssetBase, base)
	}

	var stdout, stderr bytes.Buffer
	cliErr := browsertrace.HandleCLI(req.CLIArgs, env, &stdout, &stderr)

	resp := &Response{
		CLIStdout: stdout.String(),
		CLIStderr: stderr.String(),
		CLIErr:    cliErr,
		ErrText:   errString(cliErr),
		ExitCode:  0,
	}

	ver := CacheVersion
	if v := strings.TrimSpace(browseragent.ClientVersion()); v != "" {
		if !strings.HasPrefix(v, "v") {
			ver = "v" + v
		} else {
			ver = v
		}
	}
	resp.CacheCompleteExt = browseragent.CacheComplete(ProductBrowserTrace, ver, KindExtension)
	return resp, nil
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
```
