# Scenario

**Feature**: pure `detectRuntimeBrowser` UA helper (D1–D2)

```
detectRuntimeBrowser(ua?)
  # comment: Firefox token in UA
  -> "firefox"
  # comment: Chrome / empty / other
  -> "chrome"
```

## Preconditions

- Mode `ModeReactSrc`.
- ModuleRoot resolved by root Setup.
- Helper may live in SessionPageApp.tsx or a sibling under `react/src/ui/`.

## Steps

1. Set `Mode = ModeReactSrc`.
2. Leaf sets `ReactProbe` for Firefox vs non-Firefox contract.

## Context

- Classic TDD: helper name/body absent today → RED.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeReactSrc
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	return nil
}
```
