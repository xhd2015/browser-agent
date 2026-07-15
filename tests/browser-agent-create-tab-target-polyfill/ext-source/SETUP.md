# Scenario

**Feature**: extension background implements create_tab + Target.* polyfill

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent/**/background.js
  assert create_tab job branch
  assert Target.* intercept via chrome.tabs (not raw debugger sendCommand only)
  assert Tier A methods, session window rules, tab_id identity
```

## Preconditions

- Mode is ext-source.
- ModuleRoot resolved by root Setup.
- No real browser; content asserts only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Children set ExtSourceTarget.

## Context

- Requirements E1–E5. Source/token structure style matches browser-agent-cdp-jobs ext-source.

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
