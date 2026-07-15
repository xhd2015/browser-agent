# Scenario

**Feature**: cold cache + httptest tarball → EnsureAsset fills complete cache

```
empty XDG
httptest GET .../v0.2.0/browser-agent_v0.2.0_session-page.tar.gz → fixture tar.gz
  -> EnsureAsset(...) -> dir, nil
  -> CacheComplete(browser-agent, v0.2.0, session-page) == true
```

## Preconditions

- XDG_CACHE_HOME temp (cold).
- Server serves complete session-page fixture as tar.gz.

## Steps

1. Set `DownloadOp = DownloadOpEnsureFillsCache`.
2. Set XDG temp; fixture session-page-complete.

## Context

- Happy path for P3 ensure/download.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DownloadOp = DownloadOpEnsureFillsCache
	req.XDGCacheHome = t.TempDir()
	req.DownloadFixture = FixtureSessionPageComplete
	req.DownloadProduct = ProductBrowserAgent
	req.DownloadVersion = CacheVersion
	req.DownloadKind = KindSessionPage
	return nil
}
```
