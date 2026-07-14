# Scenario

**Feature**: SessionNew system Chrome + canonical extract workflow

```
RunDaemon -> SessionNew -> EnsureCanonicalExtension + system LaunchFn + enriched stdout
ManagedChromeTestHooks.LaunchFn records argv (no OpenChromeFn injection)
```

## Preconditions

- Ephemeral daemon on `:0`; temp `BaseDir`.
- `TestHome` for canonical path assertions.
- `ManagedChromeTestHooks.LaunchFn` injected by Run (not `OpenChromeFn`).

## Steps

1. Set `Mode = session-new`.
2. Leaf sets `SessionNewOp` and optional `NoOpenChrome`.

## Context

- Never invokes agent-run.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionNew
	return nil
}
```