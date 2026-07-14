# Scenario

**Feature**: Get returns false for unknown session

```
Get(unknown) -> ok false (no pre-create)
```

## Preconditions

- GetSessionID never created in registry.

## Steps

1. Set GetSessionID to `no-such-session`.

## Context

- Empty registry except lookup attempt.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.GetSessionID = "no-such-session"
	return nil
}
```