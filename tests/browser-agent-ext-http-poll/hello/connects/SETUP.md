# Scenario

**Feature**: known session hello marks extension connected

```
Create sess-ext-hello-ok
POST /v1/ext/hello {session_id, version, features}
  -> 200 {ok:true, phase:"extension_connected"}
GET /v1/session?session=... -> extension.connected=true
```

## Preconditions

- HelloCase = connects.
- SessionID = sess-ext-hello-ok (created).

## Steps

1. Set known session id and hello identity.

## Context

- Primary attach path for Firefox HTTP poll fallback.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.HelloCase = HelloConnects
	req.SessionID = "sess-ext-hello-ok"
	req.HelloVersion = "1.0.0"
	req.HelloFeatures = []string{"browser-agent"}
	return nil
}
```
