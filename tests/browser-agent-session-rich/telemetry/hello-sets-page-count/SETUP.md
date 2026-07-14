# Scenario

**Feature**: hello telemetry sets `session_page_count` and browser product

```
Fake Extension -> hello { session_page_count: 1, browser_product: Chrome }
GET /v1/session -> count=1, browsers includes Chrome
```

## Preconditions

- `TelemetryOp = hello-sets-page-count`.

## Steps

1. Set `TelemetryOp = TelemetryOpHelloPageCount`.
2. Set `SessionID = sess-rich-hello-1`.

## Context

- Connected + count=1 + supports BA → status may also be `ready` (not asserted here).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.TelemetryOp = TelemetryOpHelloPageCount
	req.SessionID = "sess-rich-hello-1"
	return nil
}
```