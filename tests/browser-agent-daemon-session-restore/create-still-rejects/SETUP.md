# Scenario

**Feature**: Create still fails when session exists live or on disk

```
# after restore (live registry)
RestoreSessionsFromDisk -> Create(same id) -> ErrSessionExists

# disk-only without restore
seed dir -> Create(id) -> ErrSessionExists
```

## Preconditions

- Mode = create-still-rejects.
- Create path must not be used for restore registration.

## Steps

1. Set `Mode = ModeCreateStillRejects`.
2. Leaves set CreateRejectCase, CreateSessionID, Seeds.

## Context

- Restore uses register-from-disk, not Create; Create exclusivity unchanged.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeCreateStillRejects
	ensureBaseDir(t, req)
	ensureAddr(t, req)
	return nil
}
```
