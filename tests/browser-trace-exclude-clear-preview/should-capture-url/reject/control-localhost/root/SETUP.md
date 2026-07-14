# Scenario

**Feature**: reject root control URL on localhost:43759 (req #1)

```
ShouldCaptureURL("http://localhost:43759/") -> false
```

## Preconditions

- Root path on localhost control host.

## Steps

1. Set `CaptureURL = "http://localhost:43759/"`.

## Context

- Symmetric with control-ip/root.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CaptureURL = "http://localhost:43759/"
	return nil
}
```
