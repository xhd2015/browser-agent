# Scenario

**Feature**: SessionRegistry Create — register session and write artifacts

```
Test Client -> NewSessionRegistry -> Create(id)
  -> CreateSessionResult | validation error | ErrSessionExists
```

## Preconditions

- Mode is create.
- Leaf Setup sets CreateCase and SessionID (except invalid-id).

## Steps

1. Set Mode to create.
2. Ensure BaseDir and Addr.

## Context

- Duplicate guard: registry entry **or** existing session dir on disk.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCreate
	ensureBaseDir(t, req)
	ensureAddr(t, req)
	return nil
}
```