# Scenario

**Feature**: Top-level help documents `session new` and blocking `serve`

```
HandleCLI(["--help"]) + briefUsage from bare invoke
  -> session new + blocking daemon host
```

## Preconditions

- CLIHelpProbe = top-level-mentions-session-new.

## Steps

1. Set CLIHelpProbe; CLIArgs `["--help"]`.
2. Also capture briefUsage via bare `HandleCLI([])`.

## Context

- Phase 10 polish: briefUsage must name `session new` explicitly (not only `new|`).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIHelpProbe = CLIHelpProbeTopLevelSessionNew
	req.CLIArgs = []string{"--help"}
	req.CaptureBriefUsage = true
	return nil
}
```