# Scenario

**Feature**: local asset cache under XDG / ~/.cache (P2)

```
Test Client -> t.Setenv(XDG_CACHE_HOME|HOME, temp)
  -> browseragent.AssetCacheRoot / WriteAssetCache / OpenAssetCache / CacheComplete
  -> isolated cache paths; hit / miss / product keys
```

## Preconditions

- Mode is cache.
- Leaf sets `CacheOp` and any product/version/kind/fixture fields.
- Env isolation is mandatory — no real user cache.

## Steps

1. Set `Mode = ModeCache`.

## Context

- Layout: `{AssetCacheRoot}/{product}/v{version}/{kind}/…`
- Version sealed as `v0.2.0` (`CacheVersion`).
- Classic TDD for P2 — expect RED until cache APIs exist.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCache
	if req.CacheVersion == "" {
		req.CacheVersion = CacheVersion
	}
	return nil
}
```
