# Scenario

**Feature**: formatted running status shows version and extension blocks

```
RunDaemon + session -> QueryDaemonStatus -> FormatDaemonStatus -> Version + Extension markers
```

## Preconditions

- `FormatStatusOp` running-shows-version-block.

## Steps

1. Set `FormatStatusOp = FormatStatusOpRunningShowsVersionBlock`.

## Context

- Human table only (no JSON).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FormatStatusOp = FormatStatusOpRunningShowsVersionBlock
	return nil
}
```