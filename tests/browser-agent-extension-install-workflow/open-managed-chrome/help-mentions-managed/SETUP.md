# Scenario

**Feature**: top-level --help documents open-managed-chrome

```
HandleCLI(["--help"]) -> mentions open-managed-chrome; not primary open-chrome entry
```

## Preconditions

- fullHelp / briefUsage updated.

## Steps

1. Set `OpenManagedChromeOp = help-mentions-managed`.

## Context

- Operator discovers renamed command from help.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OpenManagedChromeOp = OpenManagedChromeOpHelpMentionsManaged
	return nil
}
```