# Scenario

**Feature**: session id containing `..` is rejected

```
"id=a..b" -> ValidateSessionID -> error
```

## Preconditions

- SessionID is `a..b`.

## Steps

1. Set SessionID to `a..b`.

## Context

- `..` segments must be rejected to prevent path traversal.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "a..b"
	return nil
}
```