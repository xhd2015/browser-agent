# Scenario

**Feature**: Chrome + Firefox background sources handle prepare_reconnect (markers)

```
# no browser
Test Client -> read Chrome-Ext-Browser-Agent / Firefox-Ext-Browser-Agent background.js
Test Client -> assert prepare_reconnect handler, delay_ms, schedule close, aggressive reconnect
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- Prefer `public/background.js`; accept build/src/embedded fallbacks.
- No real browser; content asserts only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Children set `ExtSourceTarget` (chrome-background | firefox-background).

## Context

- Requirement markers: `prepare_reconnect`, `delay_ms`, reconnect after close.
- Both browser shells must implement the same control-plane contract.

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
