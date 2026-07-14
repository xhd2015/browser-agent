# Scenario

**Feature**: blank window — no URL in argv

```
BuildManagedChromeArgs(dataDir, extPath, "")
  no http(s) url arg
  has user-data-dir + load-extension + new-window
```

## Preconditions

- ChromeArgsOp blank-window.

## Steps

1. Set ChromeArgsOp = ChromeArgsOpBlankWindow.

## Context

- Default URL omitted per locked decision.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ChromeArgsOp = ChromeArgsOpBlankWindow
	return nil
}
```
