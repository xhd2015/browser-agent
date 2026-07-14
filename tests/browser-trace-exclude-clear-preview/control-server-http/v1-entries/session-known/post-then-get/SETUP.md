# Scenario

**Feature**: POST entries snapshot then GET matches count and URLs (req #3)

```
Mock Extension -> POST /v1/entries {
  session_id: <live>,
  entries: [alpha, app.js],  # non-control fixture URLs
  count: 2
}
Test Client -> GET /v1/entries?session=<live>
Control Server -> count=2, URLs match fixture
```

## Preconditions

- Two HAR-like fixture entries with `https://api.example.com/v1/alpha` and
  `https://cdn.example.com/assets/app.js`.
- No clear after stage.

## Steps

1. Set `DoStagePost = true`.
2. Set `StageEntries = sampleEntries()`.
3. Set `DoClearAfterStage = false`.

## Context

- Simulates the extension’s periodic push of current in-memory entries.
- Server replaces snapshot (not required to merge incrementally).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoStagePost = true
	req.StageEntries = sampleEntries()
	req.DoClearAfterStage = false
	return nil
}
```
