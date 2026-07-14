# Scenario

**Feature**: pretty status table markers

```
RunDaemon + session -> QueryDaemonStatus -> FormatDaemonStatus -> table text
```

## Preconditions

- `FormatStatusOp` pretty-output-markers.

## Steps

1. Set `FormatStatusOp = FormatStatusOpPrettyMarkers`.

## Context

- Uses live daemon status as formatter input.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FormatStatusOp = FormatStatusOpPrettyMarkers
	return nil
}
```