# Scenario

**Feature**: serve --kill-existing always kills (Q10/Q14)

```
serve --kill-existing -> warn connected + RemoveAll orphan dirs
```

## Preconditions

- Mode `ModeKillExisting`.
- Leaf sets op-specific field.

## Steps

1. Set `Mode = ModeKillExisting`.

## Context

- See root DOCTEST for `Run` dispatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeKillExisting
	return nil
}
```
