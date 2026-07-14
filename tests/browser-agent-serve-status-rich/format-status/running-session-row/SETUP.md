# Scenario

**Feature**: formatted sessions table shows session id and phase

```
RunDaemon + session -> FormatDaemonStatus -> session row in table
```

## Preconditions

- `FormatStatusOp` running-session-row.

## Steps

1. Set `FormatStatusOp = FormatStatusOpRunningSessionRow`.

## Context

- Baseline table row content (phase7 parity).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FormatStatusOp = FormatStatusOpRunningSessionRow
	return nil
}
```