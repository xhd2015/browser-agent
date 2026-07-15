# Scenario

**Feature**: second EnsureAsset does not re-download when cache complete

```
EnsureAsset once -> cache complete (GET count = 1)
EnsureAsset again -> ok; GET count stays 1
```

## Preconditions

- XDG temp cold initially.
- httptest serves complete session-page tar.gz.
- `DownloadCallTwice = true`.

## Steps

1. Set `DownloadOp = DownloadOpSecondEnsureNoRefetch`.
2. Set XDG temp; call Ensure twice in Run.

## Context

- Verifies cache short-circuit / no redundant network.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DownloadOp = DownloadOpSecondEnsureNoRefetch
	req.XDGCacheHome = t.TempDir()
	req.DownloadFixture = FixtureSessionPageComplete
	req.DownloadCallTwice = true
	req.DownloadProduct = ProductBrowserAgent
	req.DownloadVersion = CacheVersion
	req.DownloadKind = KindSessionPage
	return nil
}
```
