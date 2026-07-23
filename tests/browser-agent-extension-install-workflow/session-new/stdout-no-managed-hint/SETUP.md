# Scenario

**Feature**: SessionNew stdout does not suggest `open-managed-chrome`

```
SessionNew -> stdout does not recommend opening a separate managed Chrome profile
```

## Preconditions

- Session URL available in stdout.

## Steps

1. Set `SessionNewOp = stdout-no-managed-hint`.

## Context

- `open-managed-chrome` is an explicit operator action, not a default next step.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpStdoutNoManagedHint
	return nil
}
```
