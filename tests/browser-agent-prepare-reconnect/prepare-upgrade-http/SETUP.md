# Scenario

**Feature**: POST /v1/admin/prepare-upgrade broadcasts prepare_reconnect and returns notified ids

```
Fake Extension hello -> POST /v1/admin/prepare-upgrade
  -> 200 { ok: true, notified: [...] }
  -> WS type=prepare_reconnect
GET /v1/admin/prepare-upgrade -> 405
```

## Preconditions

- Mode is `prepare-upgrade-http`.
- Handler registered on `NewRegistryControlHandler` mux (RED until implementer).
- Documented path: **`/v1/admin/prepare-upgrade`**.

## Steps

1. Set `Mode = ModePrepareUpgradeHTTP`.
2. Children set PrepareUpgradeCase, HTTP method, body, ConnectForHTTP.

## Context

- Phase 3 EnsureDaemon will call this endpoint before killing the old daemon.
- Path may be documented equivalently; tests pin `/v1/admin/prepare-upgrade`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModePrepareUpgradeHTTP
	return nil
}
```
