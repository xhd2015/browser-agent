# Scenario

**Feature**: HandleCLI serve --stop

```
RunDaemon (optional) + server.json
HandleCLI serve --stop -> KillExistingDaemon only -> no second RunDaemon
```

## Preconditions

- Mode `ModeStop`.
- `--no-open-chrome`, `--no-agent-run` on running-daemon leaf.

## Steps

1. Set `Mode = ModeStop`.

## Context

- `--stop` must return quickly without binding a new listener.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeStop
	return nil
}
```