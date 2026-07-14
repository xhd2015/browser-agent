# Scenario

**Feature**: expand when neither connected nor supports (req #4)

```
ShouldExpandInstallPanel(false, false) -> true
```

## Preconditions

- Connected=false, Supports=false.

## Steps

1. Set `Connected = false`.
2. Set `Supports = false`.
3. WantExpand already true from parent.

## Context

- Matches no-hello / waiting_extension mental model.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Connected = false
	req.Supports = false
	return nil
}
```
