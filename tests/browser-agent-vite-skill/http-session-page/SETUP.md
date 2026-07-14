# Scenario

**Feature**: control server serves embedded session SPA + assets (A/B)

```
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun)
Test Client -> GET /go | / | /assets/*
  -> 200 HTML/asset with product boot + install markers
```

## Preconditions

- Mode = ModeHTTP.
- Live server; no extension WS required for install guidance or boot markers.
- Embed fixture staged; server prefers embed over pure Go HTML when present.

## Steps

1. Set Mode = ModeHTTP.
2. NoOpenChrome / NoAgentRun already from root.
3. HTTPProbe set by leaf.

## Context

- Free port + temp BaseDir per leaf.
- Product markers: browser-agent, 43761, /v1/session.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeHTTP
	req.NoOpenChrome = true
	req.NoAgentRun = true
	return nil
}
```
