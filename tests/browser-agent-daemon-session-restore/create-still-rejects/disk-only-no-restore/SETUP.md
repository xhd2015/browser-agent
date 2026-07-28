# Scenario

**Feature**: Create rejects when session dir exists even without restore

```
seed sessions/sess-diskonly/ (no restore) -> Create -> ErrSessionExists
```

## Preconditions

- CreateRejectCase = disk-only-no-restore.
- Seed dir for `sess-diskonly` (meta optional; dir presence is enough for Exists/Create guard).

## Steps

1. Set CreateRejectCase, CreateSessionID, Seeds.

## Context

- Crash-recovery guard: Create does not clobber leftover dirs.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.CreateRejectCase = CreateRejectDiskOnly
	req.CreateSessionID = "sess-diskonly"
	req.Seeds = []SeedSession{
		{ID: "sess-diskonly", NoMeta: true},
	}
	return nil
}
```
