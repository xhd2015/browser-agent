# Scenario

**Feature**: GET `/v1/session` JSON status for the page poller

```
# Page poller (or Test Client) reads live session snapshot
Test Client -> GET /v1/session?session=<id>
Control Server -> JSON { session_id, phase, extension, recording, ready, hint }
```

## Preconditions

- Probe target is the session JSON API (not the HTML page).
- Session id is either the live session or an intentional unknown id (child branches).

## Steps

1. Set `Probe = ProbeV1Session` (`"v1-session"`).
2. Descendants choose known vs missing session id and staging (hello/status).

## Context

- Same-origin with `/go` so the real page can poll without CORS.
- Unknown non-empty session id must yield HTTP 404 (documented product choice).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Probe = ProbeV1Session
	return nil
}
```
