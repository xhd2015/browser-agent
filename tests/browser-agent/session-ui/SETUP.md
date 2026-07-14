# Scenario

**Feature**: session JSON snapshot and SPA shell HTML for browser-agent

```
Test Client -> browseragent.Run -> Control Server
Test Client -> GET /v1/session  -> JSON phase/extension/supports_browser_agent
Test Client -> GET /go or /     -> Vite session SPA hooks (session id, /v1/session, product port)
Optional: Fake Extension WS hello before probe
```

## Preconditions

- Mode is `session-ui`.
- No real Chrome / DOM automation — string/JSON inspection only.

## Steps

1. Set `Mode = ModeSessionUI`.
2. Children choose Probe and whether to WS hello first.

## Context

- Requirement E1–E4. Install panel expand policy for browser-trace is a separate tree.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionUI
	if req.HelloVersion == "" {
		req.HelloVersion = "1.0.0"
	}
	if req.HelloFeatures == nil {
		req.HelloFeatures = []string{"browser-agent"}
	}
	return nil
}
```
