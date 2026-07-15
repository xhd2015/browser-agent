# Scenario

**Feature**: AssetReleaseNames for v0.2.0 includes browser-agent session-page + extension

```
AssetReleaseNames("v0.2.0")
  contains browser-agent_v0.2.0_session-page.tar.gz
  contains browser-agent_v0.2.0_extension.tar.gz
  (recommended) browser-trace_v0.2.0_extension.tar.gz
```

## Preconditions

- ReleaseVersion = v0.2.0.

## Steps

1. Set `ReleaseOp = ReleaseOpAssetNames`.
2. Set `ReleaseVersion = CacheVersion` (`v0.2.0`).

## Context

- Sealed version string used by EnsureAsset / release packaging.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReleaseOp = ReleaseOpAssetNames
	req.ReleaseVersion = CacheVersion
	return nil
}
```
