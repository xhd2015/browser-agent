# Scenario

**Feature**: ready timeout emits rich stderr (stage + session URL + install hint)

```
# Short ready timeout, no hello
ReadyTimeout ~ 300–800ms, mock silent
Lifecycle Logger -> stderr:
  timeout / ready failure
  stage no_hello or hello/connect language
  session URL or /go?session=
  install path hint when known
```

## Preconditions

- Short `ReadyTimeout` (sub-second to ~1s) so the leaf stays fast.
- Default heartbeat (5s) will **not** fire — this leaf does not assert heartbeats.

## Steps

1. Set `ReadyTimeout` to 400ms.
2. Leave `ReadyHeartbeat` at 0 (product default 5s).
3. Run until ready deadline.

## Context

- Requirement scenario #3.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReadyTimeout = 400 * time.Millisecond
	req.ReadyHeartbeat = 0 // product default; no heartbeat expected in 400ms
	req.Quiet = false
	req.Verbose = false
	return nil
}
```
