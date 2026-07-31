# Scenario

**Feature**: Extension background.js self-heal on SW boot (P3)

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent/public/background.js
Test Client -> assert boot rediscover, reconnect WS, re-attach;
               empty sessions.keys() loop alone insufficient
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; filesystem reads only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Leaf sets `ExtSourceTarget`.

## Context

- Classic TDD: all four leaves **RED** under empty `sessions.keys()` boot loop
  until implementer lands tabs.query rediscover + rebind + reconnect + re-attach.
- MECE with attach-gate (P1) and eager-arm (P2) — do not re-assert set/leave or
  per-event eager triggers here.

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
