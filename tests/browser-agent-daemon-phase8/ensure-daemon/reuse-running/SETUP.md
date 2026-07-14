# Scenario

**Feature**: reuse running daemon without spawn

```
RunDaemon (pre-start) -> EnsureDaemon -> meta matches, SpawnFn not called
```

## Preconditions

- Daemon healthy at ephemeral addr before `EnsureDaemon`.
- `EnsureDaemonOp` reuse-running.

## Steps

1. Set `EnsureDaemonOp = EnsureDaemonOpReuseRunning`.

## Context

- Returned `DaemonMeta.Addr` must match pre-started daemon.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.EnsureDaemonOp = EnsureDaemonOpReuseRunning
	return nil
}```
