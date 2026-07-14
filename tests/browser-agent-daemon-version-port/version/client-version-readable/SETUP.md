# Scenario

**Feature**: client version readable

```
ClientVersion() <- go:embed VERSION.txt
CompareVersion(a,b) -> -1|0|+1
```

## Preconditions

- `VersionOp = VersionOpClientReadable`.

## Steps

1. Set `VersionOp = VersionOpClientReadable`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.VersionOp = VersionOpClientReadable
	return nil
}
```
