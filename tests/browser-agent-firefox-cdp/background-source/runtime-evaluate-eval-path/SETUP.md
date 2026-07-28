# Scenario

**Feature**: CDP Runtime.evaluate reuses Firefox eval / executeScript path

```
handleCdpJob / cdp branch
  method == "Runtime.evaluate"
    -> handleEvalJob(...)  OR  scripting.executeScript / tabs.executeScript
    # reuse Phase 2 eval path; no chrome.debugger Runtime.evaluate
```

## Preconditions

- BackgroundSourceTarget = runtime-evaluate-eval-path.

## Steps

1. Set `BackgroundSourceTarget = BgSrcRuntimeEvaluateEvalPath`.

## Context

- Markers: `Runtime.evaluate` + eval-path reuse (`handleEvalJob` and/or `executeScript`).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcRuntimeEvaluateEvalPath
	return nil
}
```
