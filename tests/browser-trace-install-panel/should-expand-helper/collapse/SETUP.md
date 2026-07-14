# Scenario

**Feature**: collapse cases — helper returns false only when connected && supports

```
ShouldExpandInstallPanel(connected, supports) -> false
# only when both inputs are true
```

## Preconditions

- Expected result is collapse (`WantExpand = false`).
- Child sets the single both-true pair.

## Steps

1. Set `WantExpand = false`.

## Context

- MECE sibling of `expand/` on expected boolean outcome.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WantExpand = false
	return nil
}
```
