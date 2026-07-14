# Scenario

**Feature**: open-managed-chrome renamed command + Chrome 137 warning

```
Operator -> HandleCLI(open-managed-chrome [url]) -> ManagedChromeTestHooks.LaunchFn
open-chrome removed; stderr WarnLoadExtensionIgnored
```

## Preconditions

- `ManagedChromeTestHooks.LaunchFn` injected by Run.
- `TestHome` for managed layout when needed.

## Steps

1. Set `Mode = open-managed-chrome`.
2. Leaf sets `OpenManagedChromeOp`.

## Context

- No real Chrome.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeOpenManagedChrome
	req.ManagedRoot = filepath.Join(t.TempDir(), "managed-chrome")
	return nil
}
```