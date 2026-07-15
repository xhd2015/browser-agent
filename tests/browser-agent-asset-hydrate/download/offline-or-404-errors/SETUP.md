# Scenario

**Feature**: EnsureAsset fails clearly on 404; cache not complete

```
httptest returns 404 for archive GET
  -> EnsureAsset -> non-nil error (download/http/404-ish)
  -> CacheComplete == false
```

## Preconditions

- XDG temp cold.
- `DownloadServe404 = true` (no successful archive body).

## Steps

1. Set `DownloadOp = DownloadOpOfflineOr404`.
2. Set XDG temp; enable 404 serving.

## Context

- Must not promote incomplete/corrupt cache as complete.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DownloadOp = DownloadOpOfflineOr404
	req.XDGCacheHome = t.TempDir()
	req.DownloadServe404 = true
	req.DownloadProduct = ProductBrowserAgent
	req.DownloadVersion = CacheVersion
	req.DownloadKind = KindSessionPage
	return nil
}
```
