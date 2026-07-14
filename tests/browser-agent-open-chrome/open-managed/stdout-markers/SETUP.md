# Scenario

**Feature**: CLI open-chrome pretty stdout

```
HandleCLI(open-chrome --root <dir>) -> managed profile markers + trailing newline
```

## Preconditions

- OpenManagedOp stdout-markers.
- ManagedChromeTestHooks.LaunchFn records argv (no real Chrome).

## Steps

1. Set OpenManagedOp = OpenManagedOpStdoutMarkers.

## Context

- Blank window CLI (no positional URL).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OpenManagedOp = OpenManagedOpStdoutMarkers
	return nil
}
```
