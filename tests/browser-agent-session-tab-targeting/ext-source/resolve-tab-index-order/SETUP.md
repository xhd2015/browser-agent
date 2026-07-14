# Scenario

**Feature**: 1-based tab_index over capturable tabs in session window

```
--tab-index N -> resolve Nth capturable tab (left-to-right) in entry.windowId
```

## Preconditions

- ExtSourceTarget = resolve-tab-index-order.

## Steps

1. Set `ExtSourceTarget = ExtSrcResolveTabIndexOrder`.

## Context

- Index includes session page tab; 1-based ordering per spec.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcResolveTabIndexOrder
	return nil
}
```