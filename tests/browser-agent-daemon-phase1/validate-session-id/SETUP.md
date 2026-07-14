# Scenario

**Feature**: ValidateSessionID accepts or rejects session ids per pattern rules

```
session id string
  -> ValidateSessionID(id) -> nil | descriptive error
```

## Preconditions

- Mode is validate-session-id.
- Leaf Setup sets concrete SessionID under test.

## Steps

1. Set Mode to validate-session-id.

## Context

- Pattern: `^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`
- Reject empty, leading non-alnum, `/`, `..`, length > 64, invalid chars.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeValidateSessionID
	return nil
}
```