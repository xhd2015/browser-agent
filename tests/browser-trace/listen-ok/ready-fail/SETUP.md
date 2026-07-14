# Scenario

**Feature**: extension never reaches recording within ready-timeout → fail

```
# Server up, but Mock Extension does not achieve recording state in time
Control Server (waiting_extension) <- (no recording status)
ready-timeout expires -> exit ≠ 0, clear extension hint
```

## Preconditions

- Ready timeout is short (sub-second) so the leaf finishes quickly.
- Complete timeout is unused or short (session fails at ready phase).

## Steps

1. Set short `ReadyTimeout`.
2. Descendants choose `ExtensionScript`: silent (`none` / no hello) vs hello without recording.

## Context

- Requirement scenarios #2 and #3.
- Error messaging should guide the user to install/enable the extension and
  `host_permissions` for `http://127.0.0.1:43759/*` (wording flexible).

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReadyTimeout = 400 * time.Millisecond
	req.CompleteTimeout = 400 * time.Millisecond
	req.StopMode = StopNone
	return nil
}
```
