# Scenario

**Feature**: serve --stop TTY confirm (Q15)

```
serve --stop + connected -> TTY [Y/n] or non-TTY warn+stop
```

## Preconditions

- Mode `ModeServeStop`.
- Leaf sets op-specific field.

## Steps

1. Set `Mode = ModeServeStop`.

## Context

- See root DOCTEST for `Run` dispatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeServeStop
	return nil
}
```
