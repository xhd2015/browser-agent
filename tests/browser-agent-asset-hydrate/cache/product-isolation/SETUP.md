# Scenario

**Feature**: browser-agent cache write does not satisfy browser-trace

```
WriteAssetCache(browser-agent, v0.2.0, session-page, complete fixture)
  -> Open/CacheComplete(browser-agent, ...) hit
  -> Open/CacheComplete(browser-trace, same version/kind) miss
```

## Preconditions

- XDG temp.
- Write only product `browser-agent`.

## Steps

1. Set `CacheOp = CacheOpProductIsolation`.
2. Set XDG temp; write fixture session-page-complete; version v0.2.0;
   kind session-page.

## Context

- Product segment isolates agent vs trace cache trees.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CacheOp = CacheOpProductIsolation
	req.XDGCacheHome = t.TempDir()
	req.CacheVersion = CacheVersion
	req.CacheKind = KindSessionPage
	req.CacheWriteFixture = FixtureSessionPageComplete
	// product set inside Run for writer vs other; leaf defaults unused for product field
	req.CacheProduct = ProductBrowserAgent
	return nil
}
```
