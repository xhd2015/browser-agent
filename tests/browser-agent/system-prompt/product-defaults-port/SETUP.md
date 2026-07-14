# Scenario

**Feature**: product default control port is 43761 (F2)

```
DefaultAddr / DefaultControlPort -> 127.0.0.1:43761 or port 43761
```

## Preconditions

- SystemOp product-defaults.
- Package exports DefaultAddr and/or DefaultControlPort.

## Steps

1. Set SystemOp to product-defaults.

## Context

- browser-trace remains 43759 (external regression); this leaf asserts agent side only.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SystemOp = SystemOpDefaults
	return nil
}
```
