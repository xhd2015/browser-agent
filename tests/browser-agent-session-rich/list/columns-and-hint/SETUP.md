# Scenario

**Feature**: session list wider columns + 0-page delete footer hint

```
hello { session_page_count: 0 }
HandleCLI session list -> Created/Pages/Browser/Status + delete hint
```

## Preconditions

- `ListOp = columns-and-hint`.
- Session reports zero pages via hello telemetry.

## Steps

1. Set `ListOp = ListOpColumnsHint`.
2. Set `SessionID = sess-rich-list-zero`.

## Context

- Footer hint for 0-page sessions suggests `session delete`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListOp = ListOpColumnsHint
	req.SessionID = "sess-rich-list-zero"
	return nil
}
```