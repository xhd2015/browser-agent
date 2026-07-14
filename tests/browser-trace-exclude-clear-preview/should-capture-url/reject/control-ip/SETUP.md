# Scenario

**Feature**: exclude `127.0.0.1:43759` (product DefaultAddr host form)

```
# Control Server product address family
ShouldCaptureURL("http://127.0.0.1:43759/…") -> false
```

## Preconditions

- Host is literal `127.0.0.1`, port `43759`, scheme `http`.
- Children refine path (`/` vs resource path).

## Steps

1. Set default `CaptureURL` to control-IP root; path leaves override.

## Context

- Product default listen address is `127.0.0.1:43759`.
- Extension agent talks to this host for hello/status/entries/preview.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Default control-IP root; with-path leaf overrides full URL.
	req.CaptureURL = "http://127.0.0.1:43759/"
	return nil
}
```
