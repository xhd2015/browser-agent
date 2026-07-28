# Scenario

**Feature**: handleJob dispatches real branches for all Phase 2 core job types

```
handleJob(msg)
  switch jobType
    case "info" | "eval" | "run" | "logs" | "screenshot" | "create_tab"
      -> real handler (not only default stub)
```

## Preconditions

- BackgroundSourceTarget = handle-job-dispatch.

## Steps

1. Set `BackgroundSourceTarget = BgSrcHandleJobDispatch`.

## Context

- String/token presence is enough (switch/if/case forms OK).
- Core set for dispatch completeness: info, eval, run, logs, screenshot, create_tab.
- CDP may remain unimplemented (Phase 3).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcHandleJobDispatch
	return nil
}
```
