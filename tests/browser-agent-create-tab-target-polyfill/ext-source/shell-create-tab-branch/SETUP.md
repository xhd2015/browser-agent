# Scenario

**Feature**: shell background has create_tab job branch (E1)

```
Read Chrome-Ext-Browser-Agent background.js
  must mention create_tab job type + chrome.tabs.create
```

## Preconditions

- ExtSourceTarget = shell-create-tab-branch.

## Steps

1. Set ExtSrcShellCreateTabBranch.

## Context

- Requirement E1. Shared create path with Target.createTarget preferred but not exclusive assert here.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcShellCreateTabBranch
	return nil
}
```
