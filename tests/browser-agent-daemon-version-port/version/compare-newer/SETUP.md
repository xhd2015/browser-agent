# Scenario

**Feature**: compare newer

```
ClientVersion() <- go:embed VERSION.txt
CompareVersion(a,b) -> -1|0|+1
```

## Preconditions

- `VersionOp = VersionOpCompareNewer`.

## Steps

1. Set `VersionOp = VersionOpCompareNewer`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.VersionOp = VersionOpCompareNewer
	req.CompareA = "0.2.0"
		req.CompareB = "0.1.0"
	return nil
}
```
