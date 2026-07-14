# Scenario

**Feature**: eval with tab_id hits background tab without focus

```
Tab1 session page; Tab2 active user; Tab3 background marker
POST eval { tab_id: backgroundTabId } -> eval_url contains BG_MARKER
```

## Preconditions

- PlaywrightOp = eval-tab-id-background.
- SessionID = sess-tab-bg.

## Steps

1. Set `PlaywrightOp = PlaywrightOpEvalTabIDBackground`.
2. Set `SessionID = sess-tab-bg`.

## Context

- Active tab is user tab; explicit tab_id pins background tab.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpEvalTabIDBackground
	req.SessionID = "sess-tab-bg"
	return nil
}
```