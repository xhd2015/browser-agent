# Scenario

**Feature**: Firefox-Ext public source contracts for connect (register / WS hello / keepalive)

```
# no real Firefox
Test Client -> read Firefox-Ext-Browser-Agent/public/contentScript.js
Test Client -> read Firefox-Ext-Browser-Agent/public/background.js
Test Client -> assert register, WebSocket hello, keepalive markers
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; filesystem reads only.
- Prefer `Firefox-Ext-Browser-Agent/public/` (accept build/ fallbacks).

## Steps

1. Set `Mode = ModeExtSource`.
2. Children set `ExtSourceTarget`.

## Context

- Phase 1 A: connect agent only (no job handlers required here).

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
