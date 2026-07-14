# Scenario

**Feature**: extension has POSTed `/v1/hello` before the session probe

```
# Extension Agent announces presence
Test Client -> POST /v1/hello { version, features? }
Control Server Session: extension.connected=true (version/features stored)
# Capability gate applied on server from hello payload
```

## Preconditions

- Live known session id.
- Hello is always posted on this branch before GET `/v1/session`.
- Exact version/features and post-hello recording are chosen by descendants.

## Steps

1. Set `DoHello = true`.
2. Leave default version/features empty for grouping; children set concrete payloads.
3. Default `DoStatusRecording = false` until a recording leaf enables it.

## Context

- After hello without recording, preferred UI phase is `extension_connected`
  (finer than staying on `waiting_extension`).
- `supports_browser_trace` depends on features + version rule, not merely connection.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoHello = true
	// Recording is opt-in on descendant leaves only.
	req.DoStatusRecording = false
	return nil
}
```
