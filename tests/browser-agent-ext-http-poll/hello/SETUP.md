# Scenario

**Feature**: `POST /v1/ext/hello` attaches extension without WebSocket

```
registry.Create(session_id)
POST /v1/ext/hello {session_id, version?, features?}
  -> 200 {ok:true, phase:"extension_connected"}
  -> GET /v1/session extension.connected=true
unknown session_id -> 404
missing session_id -> 400
```

## Preconditions

- Mode = hello.
- httptest registry handler; Create session unless leaf opts out.

## Steps

1. Set Mode to hello.
2. Leaves set HelloCase and body variants.

## Context

- Same markHello semantics as WS hello (phase + connected).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeHello
	return nil
}
```
