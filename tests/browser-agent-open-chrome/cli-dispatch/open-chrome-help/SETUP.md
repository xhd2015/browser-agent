# Scenario

**Feature**: open-chrome --help mentions managed profile

```
HandleCLI(open-chrome --help) -> managed + --root
```

## Preconditions

- CLIDispatchOp open-chrome-help.

## Steps

1. Set CLIDispatchOp = CLIDispatchOpOpenChromeHelp.

## Context

- Nil error on help (POSIX CLI convention).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIDispatchOp = CLIDispatchOpOpenChromeHelp
	return nil
}
```
