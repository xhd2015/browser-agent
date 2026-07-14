# Scenario

**Feature**: prerelease ignored

```
ClientVersion() <- go:embed VERSION.txt
CompareVersion(a,b) -> -1|0|+1
```

## Preconditions

- `VersionOp = VersionOpPrereleaseIgnored`.

## Steps

1. Set `VersionOp = VersionOpPrereleaseIgnored`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.VersionOp = VersionOpPrereleaseIgnored
	req.CompareA = "0.2.0-beta"
		req.CompareB = "0.2.0"
	return nil
}
```
