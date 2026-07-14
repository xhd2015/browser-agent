# Scenario

**Feature**: empty registry List returns no sessions

```
NewSessionRegistry -> List() -> len 0
```

## Preconditions

- ListSessionIDs unset (no pre-create).

## Steps

1. Leave ListSessionIDs empty.

## Context

- Fresh registry before any Create.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListSessionIDs = nil
	return nil
}
```