# Scenario

**Feature**: compare equal

```
ClientVersion() <- go:embed VERSION.txt
CompareVersion(a,b) -> -1|0|+1
```

## Preconditions

- `VersionOp = VersionOpCompareEqual`.

## Steps

1. Set `VersionOp = VersionOpCompareEqual`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.VersionOp = VersionOpCompareEqual
	req.CompareA = "0.2.0"
		req.CompareB = "0.2.0"
	return nil
}
```
