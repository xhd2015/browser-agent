# Scenario

**Feature**: OpenManagedChrome with injectable LaunchFn

```
OpenManagedChrome({Root, URL, LaunchFn}) -> sync ext + build args + LaunchFn(args)
HandleCLI open-chrome -> same managed path via ManagedChromeTestHooks
```

## Preconditions

- ModeOpenManaged.
- LaunchFn always injected; no real Chrome.

## Steps

1. Set Mode = ModeOpenManaged.

## Context

- Custom ManagedRoot isolates profile state per leaf.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeOpenManaged
	return nil
}
```
