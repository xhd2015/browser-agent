# Scenario

**Feature**: session window rules + tab_id identity in polyfill (E4)

```
Read background.js
  windowId / entry.windowId scope
  protect session page /go?session
  public results use tab_id (not required targetId)
```

## Preconditions

- ExtSourceTarget = shell-target-session-rules.

## Steps

1. Set ExtSrcShellTargetSessionRules.

## Context

- Requirement E4. Aligns with locked product decision: tab_id only.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcShellTargetSessionRules
	return nil
}
```
