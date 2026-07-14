# Scenario

**Feature**: --help lists session and nested side-commands (C1)

```
HandleCLI(["--help"])
  -> nil error
  -> lists "session" and nested cmds (info, eval, run, logs, screenshot, cdp)
  -> trailing \n
```

## Preconditions

- CLIKind = help.

## Steps

1. Set CLIKind help; leave CLIArgs empty so Run injects `--help`.

## Context

- Requirement C1. Top-level standalone `info` is not required if only under session.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIKind = CLIKindHelp
	req.CLIArgs = nil
	return nil
}
```
