# Scenario

**Feature**: extension sources are CDP-oriented (not pure stub comments)

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent/**/background.js
Test Client -> read browseragent/embedded/extension/background.js
  assert chrome.debugger / Runtime.evaluate / Page.captureScreenshot
  assert job type branches eval,run,logs,screenshot,cdp,info
```

## Preconditions

- Mode is ext-source.
- ModuleRoot resolved by root Setup.
- No real browser; content asserts only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Children set ExtSourceTarget.

## Context

- Requirement D1–D3.

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
