# Scenario

**Feature**: EnsureAsset request path pins version tag (not latest)

```
EnsureAsset(..., version=v0.2.0, ...)
  -> GET path contains "v0.2.0"
  -> GET path does not contain "latest"
  -> path shape includes product + kind archive name
```

## Preconditions

- XDG temp; httptest serves complete fixture.
- Version sealed `v0.2.0`.

## Steps

1. Set `DownloadOp = DownloadOpVersionPinPath`.
2. Set XDG temp; run single EnsureAsset.

## Context

- Documented URL:
  `{BaseURL}/v{version}/{product}_v{version}_{kind}.tar.gz`

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DownloadOp = DownloadOpVersionPinPath
	req.XDGCacheHome = t.TempDir()
	req.DownloadFixture = FixtureSessionPageComplete
	req.DownloadProduct = ProductBrowserAgent
	req.DownloadVersion = CacheVersion // v0.2.0
	req.DownloadKind = KindSessionPage
	return nil
}
```
