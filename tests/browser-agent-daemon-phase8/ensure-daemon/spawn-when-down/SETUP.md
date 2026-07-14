# Scenario

**Feature**: spawn daemon when control plane is down

```
no server.json -> EnsureDaemon -> SpawnFn called -> healthy + server.json
```

## Preconditions

- No pre-started daemon; `BaseDir` has no `server.json`.
- `EnsureDaemonOp` spawn-when-down.
- Harness injects `SpawnFn` that starts `RunDaemon` in background.

## Steps

1. Set `EnsureDaemonOp = EnsureDaemonOpSpawnWhenDown`.

## Context

- `SpawnFnCalled` must be true.
- Returned meta must include pid and addr for spawned daemon.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.EnsureDaemonOp = EnsureDaemonOpSpawnWhenDown
	return nil
}```
