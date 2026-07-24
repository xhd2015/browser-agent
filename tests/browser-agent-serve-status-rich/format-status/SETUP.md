# Scenario

**Feature**: `FormatDaemonStatus` rich stdout

```
QueryDaemonStatus -> FormatDaemonStatus(w, st) -> version + extension + Connected table
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

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeFormatStatus
	return nil
}
```