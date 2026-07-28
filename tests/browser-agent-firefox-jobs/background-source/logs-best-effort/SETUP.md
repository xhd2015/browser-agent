# Scenario

**Feature**: logs job is best-effort; empty entries OK with type logs

```
handleJob type=logs
  -> return { type: "logs", entries: [...] }  // may be empty / limited buffer
  -> no chrome.debugger Log.enable required
```

## Preconditions

- BackgroundSourceTarget = logs-best-effort.

## Steps

1. Set `BackgroundSourceTarget = BgSrcLogsBestEffort`.

## Context

- Soft contract: branch exists; empty entries allowed.
- Must not force Phase-3 debugger log pipeline.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcLogsBestEffort
	return nil
}
```
