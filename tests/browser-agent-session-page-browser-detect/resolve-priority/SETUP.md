# Scenario

**Feature**: `resolveInstallBrowser` priority order (P1–P5)

```
resolveInstallBrowser(forced, snap, installPath)
  # 1 forced prop
  -> forced chrome|firefox
  # 2 __BROWSER_AGENT_EXT__.browser
  -> firefox when content script says firefox
  # 3 detectRuntimeBrowser / UA Firefox
  -> firefox even if boot stamped chrome
  # 4 snap / boot / path
  -> firefox when legacy signals say so
  # 5 default
  -> chrome
```

## Preconditions

- Mode `ModeReactSrc`.
- ModuleRoot resolved by root Setup.
- Priority must put UA + content-script marker **above** chrome-stamped boot.

## Steps

1. Set `Mode = ModeReactSrc`.
2. Leaf sets `ReactProbe` for the winning-signal branch.

## Context

- Classic TDD: current code is snap/boot/path-first without EXT/UA → RED.

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
