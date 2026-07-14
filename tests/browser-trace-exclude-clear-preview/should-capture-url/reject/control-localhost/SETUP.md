# Scenario

**Feature**: exclude `localhost:43759` (approved Q1 — yes)

```
ShouldCaptureURL("http://localhost:43759/…") -> false
```

## Preconditions

- Host is literal `localhost`, port `43759`, scheme `http`.
- Children refine path variants.

## Steps

1. Set default `CaptureURL` to localhost control root; path leaves override.

## Context

- Some Chrome / OS stacks resolve or present control traffic as `localhost`
  rather than `127.0.0.1`; both must be excluded.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Default localhost control root; with-path leaf overrides full URL.
	req.CaptureURL = "http://localhost:43759/"
	return nil
}
```
