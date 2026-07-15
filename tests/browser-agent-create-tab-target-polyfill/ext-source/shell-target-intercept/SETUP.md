# Scenario

**Feature**: Target.* methods are intercepted for polyfill (E2)

```
Read background.js
  method starts with "Target." -> polyfill path
  must not only fall through to raw chrome.debugger.sendCommand for Target.*
```

## Preconditions

- ExtSourceTarget = shell-target-intercept.

## Steps

1. Set ExtSrcShellTargetIntercept.

## Context

- Requirement E2. Structural/token presence is enough (no Chrome).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcShellTargetIntercept
	return nil
}
```
