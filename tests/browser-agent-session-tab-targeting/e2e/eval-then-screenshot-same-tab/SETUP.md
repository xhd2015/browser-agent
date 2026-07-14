# Scenario

**Feature**: eval then screenshot with same tab_id both succeed

```
Tab1 session page; Tab2 user tab
POST eval { tab_id } -> ok; POST screenshot { tab_id } -> ok
```

## Preconditions

- PlaywrightOp = eval-then-screenshot-same-tab.
- SessionID = sess-tab-shot.

## Steps

1. Set `PlaywrightOp = PlaywrightOpEvalThenScreenshotSameTab`.
2. Set `SessionID = sess-tab-shot`.

## Context

- Attach reuse: same tab_id across consecutive jobs must not break screenshot.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpEvalThenScreenshotSameTab
	req.SessionID = "sess-tab-shot"
	return nil
}
```