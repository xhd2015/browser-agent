# Scenario

**Feature**: core jobs are not only Phase-1 "not implemented" stubs

```
// Phase 1 (insufficient for Phase 2 core types):
handleJob -> sendJobResult(ok=false, stub:true, "not implemented: firefox job runner…")

// Phase 2 required for info, eval, create_tab, screenshot:
handleJob -> real branches + WebExtensions APIs (tabs/scripting)
```

## Preconditions

- BackgroundSourceTarget = core-jobs-not-stub-only.

## Steps

1. Set `BackgroundSourceTarget = BgSrcCoreJobsNotStubOnly`.

## Context

- Explicit requirement leaf: jobs are **NOT** only "not implemented" stubs for
  **info / eval / create_tab / screenshot**.
- Unknown types / CDP may still return not-implemented (Phase 3).
- Helper `isPhase1JobStubOnly` encodes the RED/GREEN gate.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcCoreJobsNotStubOnly
	return nil
}
```
