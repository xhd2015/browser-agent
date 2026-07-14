# Scenario

**Feature**: expand when supports=true but not connected (req #4 edge)

```
ShouldExpandInstallPanel(false, true) -> true
```

## Preconditions

- Connected=false, Supports=true.
- Unusual in product (supports normally follows hello) but pure function must
  still expand when not connected.

## Steps

1. Set `Connected = false`.
2. Set `Supports = true`.

## Context

- Completes the truth table; guards against `return !supports` mistakes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Connected = false
	req.Supports = true
	return nil
}
```
