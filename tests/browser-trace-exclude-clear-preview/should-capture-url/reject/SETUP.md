# Scenario

**Feature**: reject capture for control-server traffic (req #1)

```
ShouldCaptureURL(control-url) -> false
# control-url host is 127.0.0.1:43759 or localhost:43759
```

## Preconditions

- Expected result is **false** (`WantCapture = false`).
- Children choose host form (IP vs localhost) and path shape.

## Steps

1. Set `WantCapture = false`.

## Context

- Sibling of `allow/` under MECE on expected boolean outcome.
- Both IP and localhost must be excluded (approved Q1).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WantCapture = false
	return nil
}
```
