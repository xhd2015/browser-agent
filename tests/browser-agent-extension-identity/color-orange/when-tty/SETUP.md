# Scenario

**Feature**: ColorOrangeIfTTY true → contains ESC / 208 (D1)

```
ColorOrangeIfTTY("…", true)
  -> contains \x1b and 38;5;208 (or equivalent orange SGR)
  -> contains original message text
```

## Preconditions

- IsTTY = true.

## Steps

1. Set IsTTY true.

## Context

- Requirement D1. Orange: `\x1b[38;5;208m … \x1b[0m`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.IsTTY = true
	req.ColorInput = "browser-agent: warning: extension identity mismatch"
	return nil
}
```
