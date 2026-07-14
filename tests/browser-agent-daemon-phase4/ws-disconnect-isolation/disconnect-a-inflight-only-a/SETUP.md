# Scenario

**Feature**: Close A WS — A inflight job fails; B unaffected

```
hello on A and B; inflight job on A; close A socket
A job ok=false (disconnect); B connected=true; B job ok=true
```

## Preconditions

- Extension on A does not auto-complete.
- Extension on B auto-completes jobs.

## Steps

1. Default harness: hang job on A, verify B after A disconnect.

## Context

- "B unaffected" = B stays connected and can run a new job successfully.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```