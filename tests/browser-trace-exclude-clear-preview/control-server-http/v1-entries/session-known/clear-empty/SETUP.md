# Scenario

**Feature**: clear captured → POST empty snapshot → GET count 0 (req #4)

```
Mock Extension -> POST /v1/entries {entries: […], count: 2}   # had data
Mock Extension -> POST /v1/entries {entries: [], count: 0}    # clear
Test Client -> GET /v1/entries?session=<live>
Control Server -> count=0, entries empty
```

## Preconditions

- Clear does **not** stop recording (product: state stays recording); this tree
  only verifies server snapshot reset via empty POST.
- Fixture first post uses sample non-control entries.

## Steps

1. Set `DoStagePost = true` with `StageEntries = sampleEntries()`.
2. Set `DoClearAfterStage = true` (empty POST after stage).

## Context

- Product: popup Clear → confirm → `clearCaptured()` → immediate empty push.
- Confirm dialog itself is out of scope.

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
