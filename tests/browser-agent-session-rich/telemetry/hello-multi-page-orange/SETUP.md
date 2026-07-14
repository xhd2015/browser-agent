# Scenario

**Feature**: hello with count=2 → `status=multiple_pages`

```
Fake Extension -> hello { session_page_count: 2 }
GET /v1/session -> status multiple_pages
```

## Preconditions

- `TelemetryOp = hello-multi-page-orange`.

## Steps

1. Set `TelemetryOp = TelemetryOpMultiPage`.
2. Set `SessionID = sess-rich-multi`.

## Context

- count>1 → `multiple_pages` (orange) even when extension connected.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.TelemetryOp = TelemetryOpMultiPage
	req.SessionID = "sess-rich-multi"
	return nil
}
```