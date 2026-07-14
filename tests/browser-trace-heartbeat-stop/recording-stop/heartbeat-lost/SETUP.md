# Scenario

**Feature**: recording silence beyond HeartbeatTimeout → heartbeat_lost (exit 0)

```
# After recording established, mock stops posting status/entries
Mock Extension -> recording status (+ optional entries) -> silence
Control Server -> HeartbeatTimeout elapses without refresh
Control Server -> save HAR from previewEntries + meta partial
browser-trace -> exit 0 + stderr warning + stdout session path\n
```

## Preconditions

- Short injectable `HeartbeatTimeout` (200ms) so the leaf finishes quickly.
- ReadyTimeout remains large enough for hello+start (seconds).
- CompleteTimeout is irrelevant once heartbeat_lost wins (still set short).
- Mock never POSTs `/v1/complete`.

## Steps

1. Set `HeartbeatTimeout = 200ms`.
2. Shorten ReadyTimeout to 3s (still enough for mock).
3. Descendants choose with-snapshot vs empty-snapshot scripts.

## Context

- Requirement scenarios #3 and #4.
- Product default HeartbeatTimeout is 10s; only tests inject short values.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HeartbeatTimeout = 200 * time.Millisecond
	req.ReadyTimeout = 3 * time.Second
	req.CompleteTimeout = 2 * time.Second
	return nil
}
```
