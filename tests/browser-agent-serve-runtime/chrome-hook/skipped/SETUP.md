# Scenario

**Feature**: NoOpenChrome=true → OpenChromeFn not called (B1)

```
Config{NoOpenChrome:true, OpenChromeFn:record} -> after health: callCount==0
```

## Preconditions

- HookExpect = skipped.
- OpenChromeFn injected by Run.

## Steps

1. Set HookExpect skipped.
2. Set NoOpenChrome true (also enforced in Run).

## Context

- Flag must short-circuit before injector and production launcher.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HookExpect = HookExpectSkipped
	req.NoOpenChrome = true
	return nil
}
```
