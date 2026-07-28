# Scenario

**Feature**: Chrome BuildExtensionShell path remains unchanged by Firefox connect work

```
stageChromePublic(ShellRoot)
  -> BuildExtensionShell(ShellRoot)
  -> Chrome-Ext-Browser-Agent/build/ (not Firefox-Ext)
```

## Preconditions

- Mode = chrome-regression.
- Children set ChromeRegressionOp and stage ShellRoot.

## Steps

1. Set `Mode = ModeChromeRegression`.

## Context

- Cheap smoke; no install-chrome CLI required for Phase 1 connect tree.

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
