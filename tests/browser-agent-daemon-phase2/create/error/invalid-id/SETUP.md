# Scenario

**Feature**: Create rejects invalid session id before registry write

```
Create("a/b") -> ValidateSessionID error (not ErrSessionExists)
```

## Preconditions

- SessionID contains slash (invalid).

## Steps

1. Set CreateCase invalid-id; SessionID `a/b`.

## Context

- No session dir or registry entry should be created.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CreateCase = CreateCaseInvalidID
	req.SessionID = "a/b"
	return nil
}
```