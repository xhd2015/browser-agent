# Scenario

**Feature**: missing cache key is incomplete / open miss

```
empty XDG cache (no WriteAssetCache)
  -> CacheComplete(browser-agent, v0.2.0, session-page) -> false
  -> OpenAssetCache(...) -> ok=false
```

## Preconditions

- XDG_CACHE_HOME temp, never written for this key.

## Steps

1. Set `CacheOp = CacheOpMissIncomplete`.
2. Set XDG temp; product/version/kind for session-page key.

## Context

- Clean miss — no error required if `ok=false` and err nil.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CacheOp = CacheOpMissIncomplete
	req.XDGCacheHome = t.TempDir()
	req.CacheProduct = ProductBrowserAgent
	req.CacheVersion = CacheVersion
	req.CacheKind = KindSessionPage
	return nil
}
```
