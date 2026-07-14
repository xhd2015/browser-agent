# Scenario

**Feature**: stop is issued but final HAR never arrives within complete-timeout

```
# Recording ok, then stop without complete
Mock Extension -> recording
stop (CLI or implied) -> Session stopping
(no POST /v1/complete before CompleteTimeout)
browser-trace -> fail; do not leave corrupt final recording.har
```

## Preconditions

- Mock reaches recording then never POSTs `/v1/complete`.
- Complete timeout is short.

## Steps

1. Set short `CompleteTimeout`.
2. Use `ExtensionScript = record-no-complete`.
3. Prefer CLI stop so the server queues `stop` and waits for complete.

## Context

- Requirement scenario #6.
- Atomic write rule: no half-written final HAR replacing a good file; missing file is OK.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtRecordNoComplete
	req.StopMode = StopCLI
	req.CompleteTimeout = 400 * time.Millisecond
	return nil
}
```
