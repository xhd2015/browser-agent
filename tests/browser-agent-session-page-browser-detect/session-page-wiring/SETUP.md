# Scenario

**Feature**: SessionPageApp wires detect + resolve into InstallGuideline (W1–W3)

```
SessionPageApp
  # define / import detectRuntimeBrowser + resolveInstallBrowser
  installBrowser = resolveInstallBrowser(...)
  -> InstallGuideline browser={installBrowser}
  # keep dual install paths
  about:debugging + chrome://extensions
```

## Preconditions

- Mode `ModeReactSrc`.
- ModuleRoot resolved by root Setup.

## Steps

1. Set `Mode = ModeReactSrc`.
2. Leaf sets `ReactProbe`.

## Context

- Wiring leaves ensure helpers are not dead code.

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
