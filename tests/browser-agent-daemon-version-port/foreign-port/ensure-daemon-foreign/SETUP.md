# Scenario

**Feature**: ensure daemon foreign

```
Foreign HTTP on control port -> serve|SessionNew fail + hint
```

## Preconditions

- `ForeignPortOp = ForeignPortOpEnsureDaemon`.

## Steps

1. Set `ForeignPortOp = ForeignPortOpEnsureDaemon`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ForeignPortOp = ForeignPortOpEnsureDaemon
	return nil
}
```
