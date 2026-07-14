# Scenario

**Feature**: `FormatDaemonStatus` pretty table output

```
QueryDaemonStatus -> FormatDaemonStatus(w, st) -> table text
```

## Preconditions

- Mode `ModeFormatStatus`.
- Leaf sets `FormatStatusOp`.

## Steps

1. Set `Mode = ModeFormatStatus`.

## Context

- Formatter only; no CLI.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeFormatStatus
	return nil
}
```