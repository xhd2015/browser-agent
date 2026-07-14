# Scenario

**Feature**: `EnsureDaemon` health reuse vs spawn

```
EnsureDaemon(cfg)
  healthy + server.json -> return meta, SpawnFn not called
  down                  -> SpawnFn called, wait healthy + server.json
```

## Preconditions

- Mode `ModeEnsureDaemon`.
- Leaf sets `EnsureDaemonOp`.

## Steps

1. Set `Mode = ModeEnsureDaemon`.

## Context

- `SpawnFn` is injectable; leaves assert call count.
- Reuse leaf pre-starts `RunDaemon` before `EnsureDaemon`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeEnsureDaemon
	return nil
}```
