# Scenario

**Feature**: WriteAssetCache then OpenAssetCache + CacheComplete hit

```
WriteAssetCache(browser-agent, v0.2.0, session-page, complete fixture)
  -> dir
OpenAssetCache(...) -> ok=true
CacheComplete(...) -> true
Open again -> ok=true (stable)
```

## Preconditions

- XDG_CACHE_HOME temp.
- Source fixture: `session-page-complete` (complete tree).

## Steps

1. Set `CacheOp = CacheOpWriteThenOpenHit`.
2. Set XDG temp, product browser-agent, kind session-page, version v0.2.0.
3. Set `CacheWriteFixture = FixtureSessionPageComplete`.

## Context

- Happy path for local cache materialization (no network).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CacheOp = CacheOpWriteThenOpenHit
	req.XDGCacheHome = t.TempDir()
	req.CacheProduct = ProductBrowserAgent
	req.CacheVersion = CacheVersion
	req.CacheKind = KindSessionPage
	req.CacheWriteFixture = FixtureSessionPageComplete
	return nil
}
```
