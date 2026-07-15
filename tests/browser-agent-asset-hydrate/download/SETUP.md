# Scenario

**Feature**: EnsureAsset downloads version-pinned archive into local cache (P3)

```
Test Client -> httptest serves {product}_v{ver}_{kind}.tar.gz
Test Client -> t.Setenv(XDG_CACHE_HOME, temp)
  -> EnsureAsset(ctx, product, version, kind, cfg{BaseURL: httptest})
  -> local complete cache dir | clear error
```

## Preconditions

- Mode is download.
- Leaf sets `DownloadOp` and server behavior flags.
- Env isolation via XDG temp — no real user cache, no GitHub.

## Steps

1. Set `Mode = ModeDownload`.
2. Default version `v0.2.0`, product browser-agent, kind session-page.

## Context

- Classic TDD for P3 — expect RED until `EnsureAsset` exists.
- Archive extract must satisfy `EmbedCompleteFS` for the kind.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeDownload
	if req.DownloadVersion == "" {
		req.DownloadVersion = CacheVersion
	}
	if req.DownloadProduct == "" {
		req.DownloadProduct = ProductBrowserAgent
	}
	if req.DownloadKind == "" {
		req.DownloadKind = KindSessionPage
	}
	return nil
}
```
