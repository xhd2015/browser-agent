# Scenario

**Feature**: RunDaemon restores disk sessions into the live registry before serving

```
seed sessions/{id}/meta.json
RunDaemon(BaseDir, ephemeral addr)
  -> GET /v1/sessions includes id
```

## Preconditions

- Mode = run-daemon-wires-restore.
- Disk seeds written before RunDaemon.
- Ephemeral `127.0.0.1:0` listen (no fixed port).

## Steps

1. Set `Mode = ModeRunDaemonWires`.
2. Default ReadyTimeout 5s.
3. Leaves set RunDaemonWireCase and Seeds.

## Context

- Implements wiring: after NewSessionRegistry, call RestoreSessionsFromDisk.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeRunDaemonWires
	ensureBaseDir(t, req)
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	return nil
}
```
