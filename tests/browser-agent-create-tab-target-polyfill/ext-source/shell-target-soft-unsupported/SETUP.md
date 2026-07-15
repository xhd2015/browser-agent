# Scenario

**Feature**: Tier B soft Target methods + Tier C unsupported polyfill error (E5)

```
Read background.js (light)
  setDiscoverTargets / setAutoAttach (no-op) optional soft
  unsupported Target.* -> polyfill unsupported product error (not -32000 only)
```

## Preconditions

- ExtSourceTarget = shell-target-soft-unsupported.

## Steps

1. Set ExtSrcShellTargetSoftUnsup.

## Context

- Requirement E5 — light coverage; secondary significance.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcShellTargetSoftUnsup
	return nil
}
```
