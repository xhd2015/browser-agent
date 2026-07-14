# Scenario

**Feature**: missing daemon version

```
ClientVersion() <- go:embed VERSION.txt
CompareVersion(a,b) -> -1|0|+1
```

## Preconditions

- `VersionOp = VersionOpMissingDaemonVer`.

## Steps

1. Set `VersionOp = VersionOpMissingDaemonVer`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.VersionOp = VersionOpMissingDaemonVer
	req.CompareA = ""
		req.ClientVersion = "0.2.0"
	return nil
}
```
