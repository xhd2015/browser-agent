# Scenario

**Feature**: WS hello telemetry updates server session page count

```
Fake Extension -> hello { session_page_count, browser_product }
GET /v1/session -> enriched telemetry fields + derived status
```

## Preconditions

- Mode is `telemetry`.
- Fake extension sends hello with telemetry payload.

## Steps

1. Set `Mode = ModeTelemetry`.
2. Leaves set `TelemetryOp`.

## Context

- Status derived from page count + connection state.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeTelemetry
	return nil
}
```