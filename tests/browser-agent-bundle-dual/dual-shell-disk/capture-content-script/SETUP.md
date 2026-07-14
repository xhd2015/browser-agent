# Scenario

**Feature**: Capture-API content script page marker (D4)

```
Chrome-Ext-Capture-API contentScript (src or public)
  __BROWSER_TRACE_EXT__ + browser-trace
```

## Preconditions

- ShellProduct = browser-trace; ShellProbe = content-script.

## Steps

1. Set ShellProduct/ShellProbe for capture content script.

## Context

- Prefer src/contentScript.js; public/ copy also OK.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ShellProduct = ShellProductCapture
	req.ShellProbe = ShellProbeContentScript
	return nil
}
```
