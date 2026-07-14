# Scenario

**Feature**: SessionNew stdout optional open-managed-chrome line in Next section

```
SessionNew -> stdout Next: browser-agent open-managed-chrome '<session URL>'
```

## Preconditions

- Session URL available in stdout.

## Steps

1. Set `SessionNewOp = stdout-open-managed-hint`.

## Context

- Optional managed profile escape hatch for operators.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpStdoutOpenManagedHint
	return nil
}
```