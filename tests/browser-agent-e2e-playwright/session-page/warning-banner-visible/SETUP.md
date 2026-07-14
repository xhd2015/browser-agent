# Scenario

**Feature**: Session warning banner visible with matching session id

```
playwright-debug -> goto /go?session=id
locator [data-browser-agent-session-warning] present
data-session-id attribute matches session id
```

## Preconditions

- `PlaywrightOp` warning-banner.
- Session id `sess-e2e-banner`.

## Steps

1. Set `PlaywrightOp = PlaywrightOpWarningBanner`.
2. Set `SessionID = "sess-e2e-banner"`.

## Context

- Script: `testdata/warning-banner.js`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpWarningBanner
	req.SessionID = "sess-e2e-banner"
	return nil
}
```