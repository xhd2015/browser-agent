# Scenario

**Feature**: serve never launches Chrome even when hook is set (B2 updated)

```
Config{NoOpenChrome:false, OpenChromeFn:record}
  -> OpenChromeFn not called (serve / Run paths do not open Chrome)
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
