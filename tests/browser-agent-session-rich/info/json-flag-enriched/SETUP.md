# Scenario

**Feature**: `session info --json` emits enriched snapshot fields

```
hello { session_page_count: 1 }
HandleCLI session info --json -> created_at, status, session_page_count
```

## Preconditions

- `InfoOp = json-flag-enriched`.
- `--json` flag set by harness.

## Steps

1. Set `InfoOp = InfoOpJSONEnriched`.
2. Set `SessionID = sess-rich-info-json`.
3. Set `JSONMode = true`.

## Context

- Machine output includes new enriched fields from sessionSnapshot.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.InfoOp = InfoOpJSONEnriched
	req.SessionID = "sess-rich-info-json"
	req.JSONMode = true
	return nil
}
```