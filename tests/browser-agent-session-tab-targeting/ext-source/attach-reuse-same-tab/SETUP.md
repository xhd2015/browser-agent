# Scenario

**Feature**: debugger attach reuse + detach on tab switch

```
Same tab_id between jobs -> reuse attach (no duplicate attach)
Different tab_id -> detach previous; serialize attach per session
```

## Preconditions

- ExtSourceTarget = attach-reuse-same-tab.

## Steps

1. Set `ExtSourceTarget = ExtSrcAttachReuseSameTab`.

## Context

- Screenshot fix: avoid double-attach race; clear error if DevTools already attached.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcAttachReuseSameTab
	return nil
}
```