# Scenario

**Feature**: cold cache + httptest tarball → EnsureAsset fills complete extension cache for browser-trace

```
empty XDG
httptest GET .../v0.2.0/browser-trace_v0.2.0_extension.tar.gz → fixture tar.gz
  -> EnsureAsset(ctx, "browser-trace", "v0.2.0", "extension", cfg) -> dir, nil
  -> CacheComplete(browser-trace, v0.2.0, extension) == true
```

## Preconditions

- XDG_CACHE_HOME temp (cold).
- Server serves complete extension fixture as tar.gz.

## Steps

1. Set `DownloadOp = DownloadOpEnsureExtension`.
2. Set XDG temp; fixture extension-complete; product browser-trace.

## Context

- Happy path: product key `browser-trace` + kind `extension` via shared EnsureAsset.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DownloadOp = DownloadOpEnsureExtension
	req.XDGCacheHome = t.TempDir()
	req.DownloadFixture = FixtureExtensionComplete
	req.DownloadProduct = ProductBrowserTrace
	req.DownloadVersion = CacheVersion
	req.DownloadKind = KindExtension
	return nil
}
```
