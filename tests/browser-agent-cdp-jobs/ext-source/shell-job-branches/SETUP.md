# Scenario

**Feature**: shell background has branches for all six job types (D2)

```
Read Chrome-Ext-Browser-Agent background.js
  must mention job types: eval, run, logs, screenshot, cdp, info
```

## Preconditions

- ExtSourceTarget = shell-job-branches.

## Steps

1. Set ExtSrcShellJobBranches.

## Context

- Requirement D2. String/token presence is enough (switch/if/case forms OK).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcShellJobBranches
	return nil
}
```
