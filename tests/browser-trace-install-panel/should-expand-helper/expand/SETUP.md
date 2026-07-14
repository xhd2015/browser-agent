# Scenario

**Feature**: expand cases — helper returns true when not (connected && supports)

```
ShouldExpandInstallPanel(connected, supports) -> true
# whenever connected && supports is false
```

## Preconditions

- Expected result is expand (`WantExpand = true`).
- Children set the three non-both input pairs.

## Steps

1. Set `WantExpand = true`.

## Context

- Sibling of `collapse/` under MECE on expected boolean outcome.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WantExpand = true
	return nil
}
```
