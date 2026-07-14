# Scenario

**Feature**: ReadDaemonMeta on missing file returns error

```
ReadDaemonMeta(missing-path) -> error
```

## Preconditions

- DaemonMetaOp is read-missing-error.
- Target path does not exist under BaseDir.

## Steps

1. Set DaemonMetaOp to read-missing-error.

## Context

- Missing or invalid file must surface read error.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DaemonMetaOp = DaemonMetaReadMissingError
	return nil
}
```