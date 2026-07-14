# Scenario

**Feature**: open-managed-chrome uses managed profile argv

```
HandleCLI(open-managed-chrome <url>) -> LaunchFn argv includes --user-data-dir
```

## Preconditions

- Optional `--root` override via `ManagedRoot`.

## Steps

1. Set `OpenManagedChromeOp = launch-has-user-data-dir`.

## Context

- Managed path distinct from system session-new launch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OpenManagedChromeOp = OpenManagedChromeOpLaunchHasUserDataDir
	return nil
}
```