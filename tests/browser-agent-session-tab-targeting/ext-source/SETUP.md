# Scenario

**Feature**: Extension background.js tab targeting contract

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent/public/background.js
Test Client -> assert tab_id window validation, tab_index order, attach lifecycle
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; filesystem reads only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Leaf sets `ExtSourceTarget`.

## Context

- Complements `browser-agent-active-tab-routing` (active-tab default) with explicit targeting.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExtSource
	return nil
}
```