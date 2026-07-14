# Scenario

**Feature**: session snapshot after serve start (no hello) → not_connected (C1)

```
serve start (no WS hello)
GET /v1/session
  -> bundled_extension {version, md5, path} set
  -> extension_match = not_connected
  -> extension.connected = false
```

## Preconditions

- No fake extension dial.

## Steps

1. Set SessionMatchKind = no-hello-not-connected.
2. DoHello false.

## Context

- Requirement C1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionMatchKind = SessionMatchNoHello
	req.DoHello = false
	return nil
}
```
