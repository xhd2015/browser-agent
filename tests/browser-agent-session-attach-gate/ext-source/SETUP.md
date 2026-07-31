# Scenario

**Feature**: Extension background.js session attach-gate + multi-tab attach (policy B)

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent/public/background.js
Test Client -> assert multi-attach keep peers, same-tab reuse+lock,
               detach-all on leave, attach gate, multi-tab recount
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; filesystem reads only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Leaf sets `ExtSourceTarget`.

## Context

- Complements tab-targeting with **session-page lifecycle** + **multi-tab attach set**.
- Classic TDD: multi-attach keep peers + detach-all **RED** under sticky switch-detach;
  gate / recount / same-tab reuse+lock may already be GREEN.

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
