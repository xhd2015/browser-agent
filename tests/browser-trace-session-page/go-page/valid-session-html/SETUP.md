# Scenario

**Feature**: GET `/go?session=valid` HTML smoke (requirement #6)

```
Control Server Session id=SessionSuffix
Test Client -> GET /go?session=<SessionSuffix>
Control Server -> HTML containing:
  - session id text
  - status UI root marker
  - /v1/session poll reference
```

## Preconditions

- Known live session id.
- No hello/status required for HTML smoke (page should render waiting state).

## Steps

1. Set `DoHello = false`, `DoStatusRecording = false`.
2. Probe `/go` after health OK.

## Context

- Status UI root: element with `data-browser-trace-status` attribute **or** a
  stable id such as `browser-trace-status` / `id="status"`.
- Inline script or fetch URL string must include `/v1/session`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoHello = false
	req.DoStatusRecording = false
	return nil
}
```
