# Scenario

**Feature**: reject root control URL on 127.0.0.1:43759 (req #1)

```
ShouldCaptureURL("http://127.0.0.1:43759/") -> false
```

## Preconditions

- Root path `/` on control IP host.

## Steps

1. Set `CaptureURL = "http://127.0.0.1:43759/"`.
2. `WantCapture` already false from ancestors.

## Context

- Covers bare origin / health-style root hits.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CaptureURL = "http://127.0.0.1:43759/"
	return nil
}
```
