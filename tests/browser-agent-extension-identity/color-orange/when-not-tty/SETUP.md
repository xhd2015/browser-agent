# Scenario

**Feature**: ColorOrangeIfTTY false → no ESC (D2)

```
ColorOrangeIfTTY(s, false)
  -> equals s (plain); no ESC sequences
```

## Preconditions

- IsTTY = false.

## Steps

1. Set IsTTY false.

## Context

- Requirement D2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.IsTTY = false
	req.ColorInput = "browser-agent: warning: extension identity mismatch"
	return nil
}
```
