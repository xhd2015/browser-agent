# Scenario

**Feature**: foreign listener fail-hard (Q3)

```
Foreign HTTP on control port -> serve|SessionNew fail + hint
```

## Preconditions

- Mode `ModeForeignPort`.
- Leaf sets op-specific field.

## Steps

1. Set `Mode = ModeForeignPort`.

## Context

- See root DOCTEST for `Run` dispatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeForeignPort
	return nil
}
```
