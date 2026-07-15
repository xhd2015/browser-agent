# Scenario

**Feature**: AssetCacheRoot respects XDG_CACHE_HOME

```
XDG_CACHE_HOME=<tmp>
  -> AssetCacheRoot()
  -> under <tmp>, path contains browser-agent and asset-cache
  -> AssetCacheDir(product,v,kind) under that root
```

## Preconditions

- Fresh temp dir assigned to XDG_CACHE_HOME.
- No need to write cache contents.

## Steps

1. Set `CacheOp = CacheOpRootRespectsXDG`.
2. Set `XDGCacheHome = t.TempDir()`.
3. Default product/version/kind for `AssetCacheDir` probe.

## Context

- Expected root:
  `filepath.Join(XDG_CACHE_HOME, "browser-agent", "asset-cache")`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CacheOp = CacheOpRootRespectsXDG
	req.XDGCacheHome = t.TempDir()
	req.CacheProduct = ProductBrowserAgent
	req.CacheVersion = CacheVersion
	req.CacheKind = KindSessionPage
	return nil
}
```
