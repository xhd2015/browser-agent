# Scenario

**Feature**: Chrome build shell and install-chrome-extension remain unchanged

```
BuildExtensionShell(ShellRoot) -> Chrome-Ext-Browser-Agent/build/
InstallChromeExtensionWithHome -> extensions/browser-agent/ + Load unpacked
```

## Preconditions

- Mode = chrome-regression.
- Smoke only; full Chrome coverage lives in sibling trees.

## Steps

1. Set Mode = ModeChromeRegression.
2. Leaf sets ChromeRegressionOp and stages as needed.

## Context

- Firefox P1 must not break Chrome defaults.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeChromeRegression
	return nil
}
```
