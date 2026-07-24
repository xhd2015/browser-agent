# Scenario

**Feature**: daemon becomes unreachable during the extension wait

```
RunDaemon -> POST /v1/sessions -> cancel daemon context
Poll GET /v1/session -> connection refused
  -> error, exit 1
```

## Preconditions

- `WaitOp = daemon-dies-during-wait`.
- Daemon is killed shortly after wait begins.
- Polling hits connection error.

## Steps

1. Set `WaitOp = WaitOpDaemonDiesDuringWait`.
2. Set `SessionID = sess-new-wait-daemon-dies`.
3. Set `WaitExtensionTimeout = 5s`.
4. Set `KillDaemonDuringWait = true`.

## Context

- Daemon is stopped after 200ms via context cancel.
- Subsequent `GET /v1/session` calls fail with connection refused.
- Error returned; exit code non-zero.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitOp = WaitOpDaemonDiesDuringWait
	req.SessionID = "sess-new-wait-daemon-dies"
	req.WaitExtensionTimeout = 5 * time.Second
	req.KillDaemonDuringWait = true
	return nil
}
```
