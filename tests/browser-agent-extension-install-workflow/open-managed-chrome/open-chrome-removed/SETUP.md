# Scenario

**Feature**: open-chrome command removed

```
HandleCLI(["open-chrome"]) -> unknown command error
```

## Preconditions

- No alias from open-chrome to open-managed-chrome.

## Steps

1. Set `OpenManagedChromeOp = open-chrome-removed`.

## Context

- Error should suggest valid commands, not open-chrome.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OpenManagedChromeOp = OpenManagedChromeOpOpenChromeRemoved
	return nil
}
```