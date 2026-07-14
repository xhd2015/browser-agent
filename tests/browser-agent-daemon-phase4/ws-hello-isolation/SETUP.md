# Scenario

**Feature**: WS hello scopes `extension.connected` to the dialed session only

```
registry {sess-p4-a, sess-p4-b}
Fake Extension -> GET /v1/ws?session=<target> hello
GET /v1/session?session=A and ?session=B -> only target connected
```

## Preconditions

- Two sessions in registry.
- Mode is `ws-hello-isolation`.

## Steps

1. Set `Mode = ModeWSHelloIsolation`.
2. Leaves set `WSHelloTarget` to A or B.

## Context

- Non-target session must remain `extension.connected=false`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeWSHelloIsolation
	return nil
}
```