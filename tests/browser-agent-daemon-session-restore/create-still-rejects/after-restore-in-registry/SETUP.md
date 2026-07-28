# Scenario

**Feature**: Create rejects id already restored into registry

```
seed + RestoreSessionsFromDisk(sess-dup) -> Create(sess-dup) -> ErrSessionExists
```

## Preconditions

- CreateRejectCase = after-restore-in-registry.
- Seed `sess-dup` with valid meta; CreateSessionID = `sess-dup`.

## Steps

1. Set CreateRejectCase, CreateSessionID, Seeds.

## Context

- Live registry membership blocks Create.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.CreateRejectCase = CreateRejectAfterRestore
	req.CreateSessionID = "sess-dup"
	req.Seeds = []SeedSession{
		{ID: "sess-dup"},
	}
	return nil
}
```
