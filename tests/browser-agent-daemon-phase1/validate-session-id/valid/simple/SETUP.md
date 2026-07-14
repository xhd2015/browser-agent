# Scenario

**Feature**: simple alphanumeric session id `my-flow` is valid

```
"id=my-flow" -> ValidateSessionID -> nil
```

## Preconditions

- SessionID is `my-flow`.

## Steps

1. Set SessionID to `my-flow`.

## Context

- Typical user-chosen flow name.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "my-flow"
	return nil
}
```