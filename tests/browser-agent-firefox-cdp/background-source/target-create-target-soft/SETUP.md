# Scenario

**Feature**: optional soft polyfill Target.createTarget → create_tab path

```
handleCdpJob / cdp branch
  method == "Target.createTarget"   # optional
    -> createTabInSession / handleCreateTabJob / tabs.create
  # If Target.createTarget is absent, leaf still PASSES (soft optional)
```

## Preconditions

- BackgroundSourceTarget = target-create-target-soft.

## Steps

1. Set `BackgroundSourceTarget = BgSrcTargetCreateTargetSoft`.

## Context

- Soft / optional: absence of Target polyfill does not fail.
- If `Target.createTarget` is present, must couple to create_tab / tabs.create.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcTargetCreateTargetSoft
	return nil
}
```
