# Scenario

**Feature**: serve flags documented in full help

```
HandleCLI(["serve", "--help"])
  -> --status (read-only) + --kill-existing + serve --session-id deprecation
```

## Preconditions

- CLIHelpProbe = serve-mentions-status-kill.

## Steps

1. Set CLIHelpProbe; CLIArgs `["serve", "--help"]`.

## Context

- Phase 7/6 behavior exists in code; phase 10 adds operator docs in `fullHelp`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIHelpProbe = CLIHelpProbeServeStatusKill
	req.CLIArgs = []string{"serve", "--help"}
	return nil
}
```