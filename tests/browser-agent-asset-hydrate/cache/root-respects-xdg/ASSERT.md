## Expected

- `AssetCacheRoot()` is non-empty.
- Root is under the leaf's `XDG_CACHE_HOME` temp dir.
- Root path (slash-normalized) contains `browser-agent` and `asset-cache`
  (e.g. `$XDG/browser-agent/asset-cache`).
- `AssetCacheDir(browser-agent, v0.2.0, session-page)` is under
  `AssetCacheRoot` and includes product, version, and kind path segments.

## Side Effects

- None required (API may create dirs lazily; path contract only).

## Errors

- Empty root, root outside XDG temp, or missing path segments fails.

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	root := resp.CacheRoot
	if strings.TrimSpace(root) == "" {
		t.Fatal("AssetCacheRoot() empty")
	}
	assertPathUnder(t, root, req.XDGCacheHome)

	slash := filepath.ToSlash(root)
	if !strings.Contains(slash, "browser-agent") {
		t.Fatalf("AssetCacheRoot %q missing browser-agent segment", root)
	}
	if !strings.Contains(slash, "asset-cache") {
		t.Fatalf("AssetCacheRoot %q missing asset-cache segment", root)
	}

	// Prefer exact join contract when possible.
	wantRoot := filepath.Join(req.XDGCacheHome, "browser-agent", "asset-cache")
	if filepath.Clean(root) != filepath.Clean(wantRoot) {
		// Allow root == .../browser-agent if AssetCacheDir still nests asset-cache;
		// but require clean wantRoot match OR root under wantRoot / wantRoot under root.
		// Strict: require equal to wantRoot for GREEN.
		t.Fatalf("AssetCacheRoot=%q want %q", root, wantRoot)
	}

	cdir := resp.CacheDir
	if strings.TrimSpace(cdir) == "" {
		t.Fatal("AssetCacheDir empty")
	}
	assertPathUnder(t, cdir, root)
	cslash := filepath.ToSlash(cdir)
	for _, seg := range []string{ProductBrowserAgent, CacheVersion, KindSessionPage} {
		if !strings.Contains(cslash, seg) {
			t.Fatalf("AssetCacheDir %q missing segment %q", cdir, seg)
		}
	}
	assertExitZero(t, resp)
}
```
