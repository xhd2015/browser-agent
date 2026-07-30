# Scenario

**Feature**: Extension background.js session attach-gate contract

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent/public/background.js
Test Client -> assert detach-on-leave, attach gate, multi-tab recount, reuse-while-open
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; filesystem reads only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Leaf sets `ExtSourceTarget`.

## Context

- Complements tab-targeting attach reuse with **session-page lifecycle** attach policy.
- Classic TDD: three policy leaves RED until implement; reuse leaf is regression-friendly.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeExtSource
	return nil
}
```
