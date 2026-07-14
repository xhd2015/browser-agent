# Scenario

**Feature**: HandleCLI open-managed-chrome succeeds

```
HandleCLI(["open-managed-chrome"]) -> exit 0; LaunchFn called once
```

## Preconditions

- Blank window (no URL).

## Steps

1. Set `OpenManagedChromeOp = cli-dispatch-ok`.

## Context

- Managed profile launch requested.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OpenManagedChromeOp = OpenManagedChromeOpCLIDispatchOK
	return nil
}
```