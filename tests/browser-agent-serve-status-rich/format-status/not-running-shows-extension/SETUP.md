# Scenario

**Feature**: formatted not-running status still shows extension block

```
no server.json -> QueryDaemonStatus -> FormatDaemonStatus -> extension without running
```

## Preconditions

- `FormatStatusOp` not-running-shows-extension.

## Steps

1. Set `FormatStatusOp = FormatStatusOpNotRunningShowsExtension`.

## Context

- Extension block present even when daemon not running.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FormatStatusOp = FormatStatusOpNotRunningShowsExtension
	return nil
}
```