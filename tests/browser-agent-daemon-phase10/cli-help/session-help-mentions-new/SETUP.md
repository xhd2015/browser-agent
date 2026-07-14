# Scenario

**Feature**: session subcommand help documents `session new`

```
HandleCLI(["session", "--help"])
  -> session new + --session-id auto-generate
```

## Preconditions

- CLIHelpProbe = session-help-mentions-new.

## Steps

1. Set CLIHelpProbe; CLIArgs `["session", "--help"]`.
2. Also capture briefUsage printed when `session` has no subcommand.

## Context

- Complements top-level briefUsage polish; bare `session` error path must name
  `session new` explicitly (not only `new|`).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIHelpProbe = CLIHelpProbeSessionHelpNew
	req.CLIArgs = []string{"session", "--help"}
	req.CaptureBriefUsage = true
	req.BriefUsageArgs = []string{"session"}
	return nil
}
```