# Scenario

**Feature**: Static contract on extension background.js tab-pick priority

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent public/background.js
Test Client -> assert pickTargetTabIdForSession active+windowId before fallback
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; filesystem reads only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Leaf sets `ExtSourceTarget`.

## Context

- Sibling `e2e/` exercises the same behavior in a real browser.

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