# Scenario

**Feature**: serve --kill-existing still works after less-flags migration

```
RunDaemon (existing) + server.json
HandleCLI serve --kill-existing -> KillExistingDaemon -> new RunDaemon -> exit 0
```

## Preconditions

- First daemon healthy with `server.json`.
- Same `--base-dir` and `--addr` for CLI serve.

## Steps

1. Leaf uses default regression Run (phase 6 cli-kill parity).

## Context

- Harness shuts down second serve via `ShutdownDaemon` so `HandleCLI` returns.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```