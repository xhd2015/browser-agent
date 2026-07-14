# Scenario

**Feature**: session id containing slash is rejected

```
"id=a/b" -> ValidateSessionID -> error
```

## Preconditions

- SessionID is `a/b`.

## Steps

1. Set SessionID to `a/b`.

## Context

- Path separators must not appear in session ids.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "a/b"
	return nil
}
```