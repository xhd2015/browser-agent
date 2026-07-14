# Scenario

**Feature**: OpenManagedChrome invokes LaunchFn once with managed argv

```
OpenManagedChrome(LaunchFn=record) -> LaunchCallCount==1
```

## Preconditions

- OpenManagedOp launch-called.
- Blank window (no URL).

## Steps

1. Set OpenManagedOp = OpenManagedOpLaunchCalled.

## Context

- Package API path (not CLI).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OpenManagedOp = OpenManagedOpLaunchCalled
	return nil
}
```
