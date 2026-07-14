# Scenario

**Feature**: Extension connects after navigating to session page

```
playwright-debug -> wait service worker -> goto /go?session=id
poll /v1/session -> extension.connected true
stdout: {"assert":"extension_connected","ok":true,"session_id":"..."}
```

## Preconditions

- `PlaywrightOp` extension-connects.
- Session id `sess-e2e-connect`.

## Steps

1. Set `PlaywrightOp = PlaywrightOpExtensionConnects`.
2. Set `SessionID = "sess-e2e-connect"`.

## Context

- Script: `testdata/extension-connects.js`.
- Harness passes argv: `baseURL`, `sessionId`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpExtensionConnects
	req.SessionID = "sess-e2e-connect"
	return nil
}
```