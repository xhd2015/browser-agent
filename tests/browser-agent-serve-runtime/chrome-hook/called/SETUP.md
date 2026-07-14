# Scenario

**Feature**: NoOpenChrome=false → OpenChromeFn called once with session URL + ext path (B2)

```
Config{NoOpenChrome:false, OpenChromeFn:record}
  -> once(sessionURL containing /go + session id, non-empty extensionInstallPath)
```

## Preconditions

- HookExpect = called.
- NoOpenChrome false; NoAgentRun true.

## Steps

1. Set HookExpect called.
2. Set NoOpenChrome false.

## Context

- sessionURL should point at session SPA (`/go`) with session id query.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HookExpect = HookExpectCalled
	req.NoOpenChrome = false
	return nil
}
```
