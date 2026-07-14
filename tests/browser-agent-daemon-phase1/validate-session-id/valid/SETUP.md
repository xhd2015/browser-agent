# Scenario

**Feature**: valid session ids pass ValidateSessionID

```
valid id (simple | dots | max length)
  -> ValidateSessionID -> nil
```

## Preconditions

- Parent mode validate-session-id.
- Leaf narrows to one valid id example.

## Steps

1. Inherit validate-session-id mode from parent.

## Context

- Valid branch: expect nil error from ValidateSessionID.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```