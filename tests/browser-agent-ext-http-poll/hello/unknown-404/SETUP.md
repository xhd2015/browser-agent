# Scenario

**Feature**: hello for unknown session returns 404

```
// no Create for target id
POST /v1/ext/hello {session_id:"no-such-session-ext-hello"}
  -> 404
```

## Preconditions

- HelloCase = unknown-404.
- OmitCreate = true (or body uses non-created id).
- HelloSessionID = no-such-session-ext-hello.

## Steps

1. Do not create the target session.
2. POST hello with unknown id.

## Context

- Mirrors other session-scoped control routes.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.HelloCase = HelloUnknown404
	req.OmitCreate = true
	req.SessionID = "sess-unused-hello-404"
	req.HelloSessionID = "no-such-session-ext-hello"
	return nil
}
```
