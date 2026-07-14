# Scenario

**Feature**: preview HTML with posted entries (req #5)

```
Mock Extension -> POST /v1/entries {entries: [alpha, app.js], count: 2}
Test Client -> GET /preview?session=<live>
Control Server -> 200 HTML
  contains at least one entry URL OR empty-table with live markers
  and poll script / entries API reference
```

## Preconditions

- Stage sample entries; no clear.
- HTML may either embed current snapshot server-side or load via client poll;
  assert allows either path as long as live markers exist and at least one
  fixture URL is present in the HTML **or** the page clearly wires `/v1/entries`
  (poll-only viewers still need a stable live marker — we require URL **or**
  both poll marker and session id).

## Steps

1. Set `DoStagePost = true` with `StageEntries = sampleEntries()`.
2. Set `DoClearAfterStage = false`.

## Context

- Primary happy-path preview leaf.

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
