# Scenario

**Feature**: packed basenames equal browseragent.AssetReleaseNames(version)

```
names := browseragent.AssetReleaseNames("v0.2.0")
go run ./script/github/release-assets --out DIR --version v0.2.0
  -> basenames(DIR) == names  (same multiset)
```

## Preconditions

- Parent pack Setup sets Mode, OutDir, Version, Args.
- Package helper `AssetReleaseNames` is already implemented (hydrate P7).

## Steps

1. Rely on parent pack Setup.
2. Version pinned to `ReleaseVersion` (`v0.2.0`).

## Context

- This leaf locks the script to the package name contract used by EnsureAsset downloads.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Version = ReleaseVersion
	// Refresh Args after version pin (parent may have already set Args).
	req.Args = []string{
		"--out", req.OutDir,
		"--version", req.Version,
	}
	return nil
}
```
