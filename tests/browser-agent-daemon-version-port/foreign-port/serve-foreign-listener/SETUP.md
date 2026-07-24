# Scenario

**Feature**: serve foreign listener

```
Foreign HTTP on control port -> serve|SessionNew fail + hint
```

## Preconditions

- `ForeignPortOp = ForeignPortOpServe`.

## Steps

1. Set `ForeignPortOp = ForeignPortOpServe`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ForeignPortOp = ForeignPortOpServe
	return nil
}
```
