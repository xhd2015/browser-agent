# Scenario

**Feature**: eval/run jobs inject page script via executeScript (no debugger)

```
handleJob type=eval|run
  -> browser.scripting.executeScript({ target, func|code })
     OR browser.tabs.executeScript(tabId, { code })
  -> return value / result (no Runtime.evaluate CDP required)
```

## Preconditions

- BackgroundSourceTarget = eval-run-execute-script.

## Steps

1. Set `BackgroundSourceTarget = BgSrcEvalRunExecuteScript`.

## Context

- Either scripting API (MV3) or tabs.executeScript is acceptable.
- Phase 2 explicitly avoids chrome.debugger / Runtime.evaluate.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcEvalRunExecuteScript
	return nil
}
```
