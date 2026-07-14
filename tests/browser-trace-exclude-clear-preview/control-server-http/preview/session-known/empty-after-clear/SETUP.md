# Scenario

**Feature**: preview after clear empty POST (req #4 + #5)

```
Mock Extension -> POST entries (non-empty)
Mock Extension -> POST empty (clear)
Test Client -> GET /preview?session=<live>
Control Server -> 200 HTML empty/cleared state
  (must not keep showing pre-clear fixture URLs as current rows)
```

## Preconditions

- Same clear contract as entries clear-empty; asserts HTML empty state.

## Steps

1. Set `DoStagePost = true` with sample entries.
2. Set `DoClearAfterStage = true`.

## Context

- Product: Clear then open Preview should show empty live view.
- Empty-state copy is flexible (empty, no requests, count 0, empty table, …).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoStagePost = true
	req.StageEntries = sampleEntries()
	req.DoClearAfterStage = true
	return nil
}
```
