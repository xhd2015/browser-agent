# Scenario

**Feature**: shell background.js references CDP debugger APIs (D1)

```
Read Chrome-Ext-Browser-Agent background.js
  must contain chrome.debugger
  must contain Runtime.evaluate
  must contain Page.captureScreenshot
```

## Preconditions

- ExtSourceTarget = shell-cdp-tokens.

## Steps

1. Set ExtSrcShellCDPTokens.

## Context

- Requirement D1. Production shell has fuller CDP implementation.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcShellCDPTokens
	return nil
}
```
