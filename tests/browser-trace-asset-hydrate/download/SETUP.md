# Scenario

**Feature**: EnsureAsset for product browser-trace extension (P6)

```
Test Client -> httptest serves browser-trace_v{ver}_extension.tar.gz
Test Client -> t.Setenv(XDG_CACHE_HOME, temp)
  -> EnsureAsset(ctx, "browser-trace", version, "extension", cfg{BaseURL: httptest})
  -> local complete cache dir | clear error
```

## Preconditions

- Mode is download.
- Leaf sets `DownloadOp` and fixture/server behavior.
- Env isolation via XDG temp — no real user cache, no GitHub.

## Steps

1. Set `Mode = ModeDownload`.
2. Default version `v0.2.0`, product browser-trace, kind extension.

## Context

- Reuses `browseragent.EnsureAsset`; product string isolates cache under
  `…/asset-cache/browser-trace/…`.
- Archive extract must satisfy `EmbedCompleteFS` for extension.

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
		req.DownloadProduct = ProductBrowserTrace
	}
	if req.DownloadKind == "" {
		req.DownloadKind = KindExtension
	}
	return nil
}
```
