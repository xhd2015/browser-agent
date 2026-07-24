# Scenario

**Feature**: open-managed-chrome renamed command + Chrome 137 warning

```
Operator -> WithManagedChromeHooks -> HandleCLI(open-managed-chrome [url]) -> LaunchFn
open-chrome removed; stderr WarnLoadExtensionIgnored
```

## Preconditions

- `inject.WithManagedChromeHooks` installs `LaunchFn` for the HandleCLI call only.
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

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeOpenManagedChrome
	req.ManagedRoot = filepath.Join(t.TempDir(), "managed-chrome")
	return nil
}
```