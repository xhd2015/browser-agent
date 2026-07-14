# Scenario

**Feature**: open-managed-chrome stderr Chrome 137 --load-extension warning (F1)

```
HandleCLI(open-managed-chrome <url>) -> stderr WarnLoadExtensionIgnored
```

## Preconditions

- Warning emitted before or after launch attempt.

## Steps

1. Set `OpenManagedChromeOp = stderr-chrome137-warning`.

## Context

- Operators directed to Load unpacked from canonical path.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OpenManagedChromeOp = OpenManagedChromeOpStderrChrome137Warn
	return nil
}
```