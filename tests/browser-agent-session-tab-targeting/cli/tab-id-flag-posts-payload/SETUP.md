# Scenario

**Feature**: --tab-id posts tab_id in job envelope

```
RunDaemon -> fake extension WS
HandleCLI session eval --tab-id 216771574 -> WS job payload tab_id=216771574
```

## Preconditions

- CLIOp = tab-id-flag-posts-payload.
- TabID = 216771574 (stable doctest value).

## Steps

1. Set `CLIOp = CLIOpTabIDPostsPayload`.
2. Set `TabID = 216771574`.
3. Set `SessionID = sess-tab-id-payload`.

## Context

- Observes first WS job envelope `tab_id` field (top-level in payload).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpTabIDPostsPayload
	req.TabID = 216771574
	req.SessionID = "sess-tab-id-payload"
	return nil
}
```