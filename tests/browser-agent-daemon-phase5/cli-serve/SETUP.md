# Scenario

**Feature**: HandleCLI serve flag behavior

```
HandleCLI(serve [flags]) -> RunDaemon or compat Run
serve --session-id -> stderr deprecation warning
```

## Preconditions

- Mode `ModeCLIServe`.
- Leaf sets `CLISessionID` and ephemeral `--addr`.

## Steps

1. Set `Mode = ModeCLIServe`.

## Context

- Deprecation leaf polls stderr early; does not wait for blocking serve exit.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLIServe
	return nil
}
```