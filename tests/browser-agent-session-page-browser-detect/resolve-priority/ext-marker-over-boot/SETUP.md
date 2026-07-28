# Scenario

**Feature**: `__BROWSER_AGENT_EXT__.browser` beats chrome-stamped boot (P2)

```
resolveInstallBrowser
  # window.__BROWSER_AGENT_EXT__.browser === "firefox"
  -> "firefox"
  # even when snap/boot stamped chrome
```

## Preconditions

- Mode already `ModeReactSrc` from parent.
- Content-script marker is second priority (after forced).

## Steps

1. Set `ReactProbe = ReactProbeResolveExtOverBoot`.

## Context

- Key override when extension injected marker is present before connect UX settles.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeResolveExtOverBoot
	return nil
}
```
