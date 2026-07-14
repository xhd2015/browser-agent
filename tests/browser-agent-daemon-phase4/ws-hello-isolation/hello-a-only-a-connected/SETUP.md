# Scenario

**Feature**: Hello on session A — only A connected

```
WS hello on sess-p4-a
GET snapshots: A connected+supports; B disconnected
```

## Preconditions

- `WSHelloTarget = A`.

## Steps

1. Set `WSHelloTarget = WSHelloTargetA`.

## Context

- B must not inherit A's hello state.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WSHelloTarget = WSHelloTargetA
	return nil
}
```