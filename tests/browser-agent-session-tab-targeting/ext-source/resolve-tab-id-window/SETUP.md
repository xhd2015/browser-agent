# Scenario

**Feature**: background validates explicit tab_id within session window

```
Job tab_id + entry.windowId -> pick/resolve target tab in session window only
```

## Preconditions

- ExtSourceTarget = resolve-tab-id-window.

## Steps

1. Set `ExtSourceTarget = ExtSrcResolveTabIDWindow`.

## Context

- Rejects or ignores tab_id outside session window.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcResolveTabIDWindow
	return nil
}
```