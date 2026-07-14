# Scenario

**Feature**: delete cleans disk-only session dirs without registry entry

```
RunDaemon (for addr resolve)
MkdirAll sessions/id only (no POST /v1/sessions)
HandleCLI session delete -> dir removed; exit 0
```

## Preconditions

- Mode `disk-only`; session dir seeded on disk only.

## Steps

1. Set `Mode = ModeDiskOnly`.

## Context

- `Exists(id)` true via disk; registry has no live session object.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeDiskOnly
	return nil
}
```