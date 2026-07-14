# Scenario

**Feature**: SessionNew uses managed Chrome instead of legacy openChrome

```
SessionNew -> OpenManagedChrome -> ManagedChromeTestHooks.LaunchFn(argv)
```

## Preconditions

- ModeSessionNewIntegration.
- Ephemeral daemon on loopback `:0`.
- ManagedChromeTestHooks.LaunchFn injected by Run.

## Steps

1. Set Mode = ModeSessionNewIntegration.

## Context

- Does not override OpenChromeFn; exercises production managed path.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionNewIntegration
	return nil
}
```
