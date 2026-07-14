# Scenario

**Feature**: 64-character valid session id at max length passes

```
"id=<64 chars>" -> ValidateSessionID -> nil
```

## Preconditions

- SessionID is exactly 64 characters, starting with alphanumeric.

## Steps

1. Set SessionID to a 64-char string: `a` + 63 `b` chars.

## Context

- Upper bound of allowed length is 64 inclusive.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "a" + strings.Repeat("b", 63)
	return nil
}
```