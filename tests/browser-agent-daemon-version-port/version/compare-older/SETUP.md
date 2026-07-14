# Scenario

**Feature**: compare older

```
ClientVersion() <- go:embed VERSION.txt
CompareVersion(a,b) -> -1|0|+1
```

## Preconditions

- `VersionOp = VersionOpCompareOlder`.

## Steps

1. Set `VersionOp = VersionOpCompareOlder`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.VersionOp = VersionOpCompareOlder
	req.CompareA = "0.1.0"
		req.CompareB = "0.2.0"
	return nil
}
```
