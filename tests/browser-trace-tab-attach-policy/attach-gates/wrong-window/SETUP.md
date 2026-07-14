# Scenario

**Feature**: do not attempt attach for tabs outside the pinned window (#6)

```
URL capturable; recording=true; windowMatch=false; alreadyAttached=false
IsCapturableTabURL -> true
ShouldAttemptAttach -> false
```

## Preconditions

- Tab lives in a different Chrome window than `targetWindowId`.
- Other gates remain open; only `WindowMatch` is false.

## Steps

1. Set `WindowMatch = false`.

## Context

- browser-trace pins one capture window; other windows must not be attached.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WindowMatch = false
	return nil
}
```
