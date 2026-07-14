# Scenario

**Feature**: extension source files implement per-session WS + tab routing (P1–P4)

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent public/background.js + contentScript.js
Test Client -> assert per-session WS URL, sessions map, register, job routing
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; content asserts only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Children set `ExtSourceTarget`.

## Context

- Requirement P1–P4. Background and content script contracts for multi-session daemon.

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