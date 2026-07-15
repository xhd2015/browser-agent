# Scenario

**Feature**: Tier A Target.* polyfill methods present (E3)

```
Read background.js
  Target.createTarget, Target.closeTarget, Target.activateTarget,
  Target.getTargets, Target.getTargetInfo
  + chrome.tabs create/remove/update/query/get
```

## Preconditions

- ExtSourceTarget = shell-target-tier-a.

## Steps

1. Set ExtSrcShellTargetTierA.

## Context

- Requirement E3 (must work methods).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcShellTargetTierA
	return nil
}
```
