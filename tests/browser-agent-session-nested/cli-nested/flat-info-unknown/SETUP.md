# Scenario

**Feature**: flat info is not a side-command handler (C3)

```
HandleCLI(["info"], empty env)
  -> non-nil error (unknown / brief)
  -> NOT a successful session handler
```

## Preconditions

- Flat top-level `info` (no `session` parent).

## Steps

1. Set CLIKind flat-info-unknown.

## Context

- Requirement C3. Complete refactor removes flat aliases.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIKind = CLIKindFlatInfoUnknown
	req.CLIArgs = nil
	req.CLIEnv = map[string]string{}
	return nil
}
```
