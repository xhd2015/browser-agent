# Scenario

**Feature**: ColorOrangeIfTTY — orange ANSI when isTTY, plain otherwise

```
Test Client -> ColorOrangeIfTTY(s, isTTY)
  isTTY=true  -> wraps with ESC[38;5;208m … ESC[0m
  isTTY=false -> returns s unchanged (no ESC)
```

## Preconditions

- Mode = color-orange.
- Leaves set IsTTY.

## Steps

1. Set Mode to color-orange.
2. Default ColorInput warning-like string.

## Context

- Requirement G5 / scenarios D1–D2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeColorOrange
	if req.ColorInput == "" {
		req.ColorInput = "browser-agent: warning: extension identity mismatch"
	}
	return nil
}
```
