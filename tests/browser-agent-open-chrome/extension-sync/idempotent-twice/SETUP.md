# Scenario

**Feature**: second EnsureManagedExtension is idempotent

```
EnsureManagedExtension x2 -> same path + same version
```

## Preconditions

- ExtensionSyncOp idempotent-twice.

## Steps

1. Set ExtensionSyncOp = ExtensionSyncOpIdempotentTwice.

## Context

- Same embedded version must not create a second directory.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionSyncOp = ExtensionSyncOpIdempotentTwice
	return nil
}
```
