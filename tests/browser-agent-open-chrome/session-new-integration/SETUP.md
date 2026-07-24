# Scenario

**Feature**: SessionNew opens system Chrome (argv shape; no real browser)

```
SessionNew(OpenChromeFn=record, NoWait) -> BuildChromeArgs(url, "") once
```

## Preconditions

- ModeSessionNewIntegration.
- Ephemeral daemon on loopback `:0`.
- `OpenChromeFn` records system-chrome argv (`--new-window` + URL; no user-data-dir).
- `NoWait` so leaves do not poll 30s for a real extension.

## Steps

1. Set Mode = ModeSessionNewIntegration.

## Context

- Per-config `OpenChromeFn` (not global `ManagedChromeTestHooks`) — parallel-safe.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionNewIntegration
	return nil
}
```
