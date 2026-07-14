# Scenario

**Feature**: session info with extension disconnected is control-only (no fake tabs) (B1)

```
serve session (no WS hello)
HandleCLI session info --session-id --addr
  -> JSON control: phase waiting* / extension.connected false
  -> no fabricated tabs of real pages
  -> browser unavailable signal (browser null | browser_error | message)
  -> trailing \n
```

## Preconditions

- Live serve; FakeExtension false.
- SessionInfoKind = control-only-disconnected.
- Nested `session info` only.

## Steps

1. Set SessionInfoKind control-only-disconnected.
2. Leave CLIArgs empty so Run builds argv with live addr/session.
3. Increase MaxDispatchWait for HTTP.

## Context

- Requirement B1. Agents must not invent tab inventory from other sources.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionInfoKind = SessionInfoControlOnlyDisconnected
	req.FakeExtension = false
	req.CLIArgs = nil
	req.MaxDispatchWait = 10 * time.Second
	return nil
}
```
