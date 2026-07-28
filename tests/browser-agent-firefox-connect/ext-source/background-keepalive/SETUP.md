# Scenario

**Feature**: Firefox background keepalive/reconnect is real (not no-op-only 1min alarm)

```
Background while WS open
  -> keepalive timer sends ping (or equivalent)
  -> dead sockets: reconnect / alarm reconnects
  -> ping → pong handling acceptable
```

## Preconditions

- ExtSourceTarget = background-keepalive.
- Current P1 stub only creates a no-op alarm — must go RED until real keepalive lands.

## Steps

1. Set `ExtSourceTarget = ExtSrcBackgroundKeepalive`.

## Context

- Empty `// no-op keepalive` alarm body is **not** sufficient.
- Chrome pattern: `startKeepalive` + `setInterval` ping + reconnect alarm is the model.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcBackgroundKeepalive
	return nil
}
```
