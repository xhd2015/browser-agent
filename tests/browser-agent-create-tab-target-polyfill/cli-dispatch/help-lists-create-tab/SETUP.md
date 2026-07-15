# Scenario

**Feature**: --help lists session create-tab (A1)

```
HandleCLI(["--help"], empty env)
  -> nil error
  -> help lists create-tab (nested under session)
  -> trailing \n
```

## Preconditions

- DispatchKind = help.
- No session / server required.

## Steps

1. Set DispatchKind to DispatchHelp.
2. Leave CLIArgs empty so Run injects `--help`.

## Context

- Requirement A1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DispatchKind = DispatchHelp
	req.CLIArgs = nil
	return nil
}
```
