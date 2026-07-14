# Scenario

**Feature**: session id longer than 64 characters is rejected

```
"id=<65 chars>" -> ValidateSessionID -> error
```

## Preconditions

- SessionID is exactly 65 characters.

## Steps

1. Set SessionID to `a` + 64 `b` chars (length 65).

## Context

- Maximum allowed length is 64.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "a" + strings.Repeat("b", 64)
	return nil
}
```