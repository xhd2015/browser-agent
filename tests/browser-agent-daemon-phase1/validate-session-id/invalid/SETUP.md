# Scenario

**Feature**: invalid session ids are rejected by ValidateSessionID

```
invalid id (empty | leading dash | slash | dotdot | too long)
  -> ValidateSessionID -> descriptive error
```

## Preconditions

- Parent mode validate-session-id.
- Leaf narrows to one invalid id example.

## Steps

1. Inherit validate-session-id mode from parent.

## Context

- Invalid branch: expect non-nil error from ValidateSessionID.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```