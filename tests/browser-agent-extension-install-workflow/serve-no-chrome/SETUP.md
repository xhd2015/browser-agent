# Scenario

**Feature**: serve / RunDaemon never opens Chrome

```
HandleCLI(serve) -> RunDaemon; LaunchFn/OpenChromeFn hooks never called
serve --help omits --no-open-chrome
```

## Preconditions

- Hooks installed before serve dispatch.
- Plain serve (no deprecated --session-id).

## Steps

1. Set `Mode = serve-no-chrome`.
2. Leaf sets `ServeNoChromeOp`.

## Context

- Chrome open moves to session new only.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeServeNoChrome
	return nil
}
```