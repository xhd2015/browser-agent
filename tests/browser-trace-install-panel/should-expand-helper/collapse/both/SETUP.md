# Scenario

**Feature**: collapse when connected and supports (req #4)

```
ShouldExpandInstallPanel(true, true) -> false
```

## Preconditions

- Connected=true, Supports=true.
- Only truth-table cell that collapses.

## Steps

1. Set `Connected = true`.
2. Set `Supports = true`.
3. WantExpand already false from parent.

## Context

- Parity with `go-html/hello-supports` collapsed default.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Connected = true
	req.Supports = true
	return nil
}
```
