# Scenario

**Feature**: Get returns true for registered session

```
Create(id) -> Get(id) -> ok true
```

## Preconditions

- Session pre-created with same id as Get target.

## Steps

1. Set SessionID and GetSessionID to `get-hit`.

## Context

- Registry holds live session after Create.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "get-hit"
	req.GetSessionID = "get-hit"
	return nil
}
```