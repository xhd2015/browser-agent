# Scenario

**Feature**: Hello on session B — only B connected

```
WS hello on sess-p4-b
GET snapshots: B connected+supports; A disconnected
```

## Preconditions

- `WSHelloTarget = B`.

## Steps

1. Set `WSHelloTarget = WSHelloTargetB`.

## Context

- A must not inherit B's hello state.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WSHelloTarget = WSHelloTargetB
	return nil
}
```