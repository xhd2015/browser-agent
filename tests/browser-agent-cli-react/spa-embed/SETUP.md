# Scenario

**Feature**: session SPA HTML from control server embed

```
Test Client -> browseragent.Run(NoOpenChrome)
Test Client -> GET /go (or /) ?session=
  -> HTML with React root + product boot hooks + install markers
```

## Preconditions

- ModeSPAEmbed; live server; no WS required for install guidance.
- Embedded session-page dist (or server-generated product HTML) available.

## Steps

1. Set Mode = ModeSPAEmbed.
2. NoOpenChrome / NoAgentRun already from root.
3. SPAProbe set by leaf.

## Context

- Prefer /go; harness falls back to /.
- Product markers: browser-agent, 43761, /v1/session.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSPAEmbed
	req.NoOpenChrome = true
	req.NoAgentRun = true
	return nil
}
```
