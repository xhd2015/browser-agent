# Scenario

**Feature**: Extension background.js session eager auto-attach (P2)

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent/public/background.js
Test Client -> assert auto-attach on /go register, create_tab,
               same-window navigate; skip non-capturable; window scope
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; filesystem reads only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Leaf sets `ExtSourceTarget`.

## Context

- Classic TDD: all five leaves **RED** under job-only attach until implementer
  lands eager call sites.
- MECE with attach-gate P1 leaves (set/gate/leave not re-asserted here).

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
