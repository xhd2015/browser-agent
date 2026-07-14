# Scenario

**Feature**: Control Server binds successfully; session lifecycle continues

```
# Listen succeeds on free Addr
browser-trace.Run -> Control Server listening on 127.0.0.1:port
GET /v1/health -> 200

# Mock Extension may or may not connect depending on branch
Mock Extension ?-> /v1/hello, /v1/commands, /v1/status, /v1/complete
```

## Preconditions

- `OccupyAddr` is false (port free for browser-trace).
- `Addr` left empty so `Run` selects a free loopback port.

## Steps

1. Ensure bind-success defaults: no port occupation.
2. Descendants choose ready/complete outcomes via `ExtensionScript` and timeouts.

## Context

- All leaves under this node expect the server process to accept the listen call.
- Ready and complete deadlines still apply after a successful bind.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OccupyAddr = false
	// Addr empty → Run allocates free port.
	req.Addr = ""
	return nil
}
```
