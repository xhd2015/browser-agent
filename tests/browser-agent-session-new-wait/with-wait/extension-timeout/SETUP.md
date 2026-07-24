# Scenario

**Feature**: no extension connects within the timeout period

```
No extension connects
Poll GET /v1/session?session=ID -> extension.connected=false throughout
  -> timeout warning on stderr, exit 0, stdout still has session output
```

## Preconditions

- `WaitOp = extension-timeout`.
- No extension connects at all.
- Timeout is short (3s) for fast test feedback.

## Steps

1. Set `WaitOp = WaitOpExtensionTimeout`.
2. Set `SessionID = sess-new-wait-timeout`.
3. Set `WaitExtensionTimeout = 3s` for fast feedback.
4. Do NOT set `ConnectExtension = true`.

## Context

- Polling happens every 500ms, so 3s gives ~6 poll cycles.
- On timeout: stderr warning, stdout session output, exit 0.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitOp = WaitOpExtensionTimeout
	req.SessionID = "sess-new-wait-timeout"
	req.WaitExtensionTimeout = 3 * time.Second
	return nil
}
```
