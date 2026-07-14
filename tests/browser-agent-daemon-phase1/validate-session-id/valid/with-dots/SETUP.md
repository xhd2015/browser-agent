# Scenario

**Feature**: session id with dots `a.b.c` is valid

```
"id=a.b.c" -> ValidateSessionID -> nil
```

## Preconditions

- SessionID is `a.b.c`.

## Steps

1. Set SessionID to `a.b.c`.

## Context

- Dots are allowed in the middle of session ids.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "a.b.c"
	return nil
}
```