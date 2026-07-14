# Scenario

**Feature**: formatted sessions table includes Connected column

```
RunDaemon + session -> FormatDaemonStatus -> Connected header + yes/no
```

## Preconditions

- `FormatStatusOp` running-shows-connected-column.

## Steps

1. Set `FormatStatusOp = FormatStatusOpRunningShowsConnectedColumn`.

## Context

- New session without extension hello → `no`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FormatStatusOp = FormatStatusOpRunningShowsConnectedColumn
	return nil
}
```