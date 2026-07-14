# Scenario

**Feature**: phase 6 kill-existing regression

```
RunDaemon (first)
HandleCLI serve --kill-existing -> kills first -> starts second -> ShutdownDaemon -> exit 0
```

## Preconditions

- Mode `ModeRegression`.
- `--no-open-chrome`, `--no-agent-run` for fast start.

## Steps

1. Set `Mode = ModeRegression`.

## Context

- Ensures new serve parsing does not break existing `--kill-existing` path.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeRegression
	return nil
}
```