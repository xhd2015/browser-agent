# Scenario

**Feature**: CLI serve --kill-existing

```
RunDaemon (existing) + server.json
HandleCLI serve --kill-existing -> KillExistingDaemon -> new RunDaemon
```

## Preconditions

- Mode `ModeCLIKill`.
- `--no-open-chrome`, `--no-agent-run` for fast start.

## Steps

1. Set `Mode = ModeCLIKill`.

## Context

- Harness shuts down second serve via `ShutdownDaemon` so `HandleCLI` returns.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLIKill
	return nil
}
```