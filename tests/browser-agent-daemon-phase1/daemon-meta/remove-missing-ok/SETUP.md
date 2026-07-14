# Scenario

**Feature**: RemoveDaemonMeta on absent file returns nil

```
RemoveDaemonMeta(missing-path) -> nil
```

## Preconditions

- DaemonMetaOp is remove-missing-ok.
- Target path does not exist under BaseDir.

## Steps

1. Set DaemonMetaOp to remove-missing-ok.

## Context

- Idempotent remove: already absent is not an error.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DaemonMetaOp = DaemonMetaRemoveMissingOK
	return nil
}
```