# Scenario

**Bug**: eval job returned session page URL when user tab was active in session window

```
playwright-debug -> Tab1 goto /go?session=S (extension connects)
                 -> Tab2 goto example.com/?LOOP_MARKER=active-tab-routing, bringToFront
POST /v1/jobs info  -> active tab has LOOP_MARKER
POST /v1/jobs eval  -> value.url contains LOOP_MARKER (not /go?session=)
stdout: {"assert":"active_tab_routing","ok":true,...}
```

## Preconditions

- `PlaywrightOp` = eval-on-active-user-tab.
- Session id `sess-active-tab-routing`.

## Steps

1. Set `PlaywrightOp = PlaywrightOpEvalOnActiveUserTab`.
2. Set `SessionID = "sess-active-tab-routing"`.

## Context

- Script: `testdata/active-tab-routing.js` (ported from `script/debug/active-tab-routing/testdata/`).
- Harness passes argv: `baseURL`, `sessionId`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpEvalOnActiveUserTab
	req.SessionID = "sess-active-tab-routing"
	return nil
}
```