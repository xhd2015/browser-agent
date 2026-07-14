# Scenario

**Feature**: allow capture for non-control traffic (req #2)

```
ShouldCaptureURL(normal-url) -> true
# host is not 127.0.0.1:43759 / localhost:43759
```

## Preconditions

- Expected result is **true** (`WantCapture = true`).
- Children set concrete non-control URLs.

## Steps

1. Set `WantCapture = true`.

## Context

- Sibling of `reject/` under MECE on expected boolean outcome.
- Includes normal HTTPS and same-loopback **different port** (must still capture).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WantCapture = true
	return nil
}
```
