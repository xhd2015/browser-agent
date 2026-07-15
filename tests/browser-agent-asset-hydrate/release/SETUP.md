# Scenario

**Feature**: release archive basenames matching EnsureAsset URL shape (P7)

```
browseragent.AssetReleaseNames(version)
  -> []string of {product}_v{version}_{kind}.tar.gz
```

## Preconditions

- Mode is release.
- Leaf sets ReleaseOp and version.

## Steps

1. Set `Mode = ModeRelease`.

## Context

- Classic TDD — RED until AssetReleaseNames exists.
- Names must match EnsureAsset: `{BaseURL}/v{ver}/{product}_v{ver}_{kind}.tar.gz`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeRelease
	if req.ReleaseVersion == "" {
		req.ReleaseVersion = CacheVersion
	}
	return nil
}
```
