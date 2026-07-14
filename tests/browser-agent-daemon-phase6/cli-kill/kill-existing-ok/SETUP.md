# Scenario

**Feature**: serve --kill-existing operator path

```
RunDaemon (first)
HandleCLI serve --kill-existing -> kills first -> starts second -> ShutdownDaemon -> exit 0
```

## Preconditions

- First daemon healthy with `server.json`.
- Same `--base-dir` and `--addr` for CLI serve.

## Steps

1. Leaf uses default `ModeCLIKill` from parent (no extra fields).

## Context

- Pretty stderr should mention kill/shutdown/existing/stopped.
- Final `HandleCLI` return value must be nil (exit 0).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```