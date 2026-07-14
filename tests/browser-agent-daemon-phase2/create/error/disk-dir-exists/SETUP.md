# Scenario

**Feature**: Create rejects when session dir already on disk (crash recovery)

```
MkdirAll session dir -> Create(id) -> ErrSessionExists
```

## Preconditions

- SessionID `disk-leftover`.
- SeedSessionDirBefore true.

## Steps

1. Set CreateCase disk-dir-exists.
2. Enable SeedSessionDirBefore.

## Context

- Dir exists but id not in registry (simulates crash recovery).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CreateCase = CreateCaseDiskDirExists
	req.SessionID = "disk-leftover"
	req.SeedSessionDirBefore = true
	return nil
}
```