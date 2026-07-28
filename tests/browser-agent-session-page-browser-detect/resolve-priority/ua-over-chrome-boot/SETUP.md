# Scenario

**Feature**: Firefox UA beats chrome-stamped snap/boot/path (P3) — primary bug fix

```
resolveInstallBrowser
  # no forced, no EXT (or EXT absent)
  # detectRuntimeBrowser() / navigator UA Firefox/
  -> "firefox"
  # even when snap.browser / boot / path stamped chrome
```

## Preconditions

- Mode already `ModeReactSrc` from parent.
- Primary Phase 1 fix: do not prefer wrong boot meta over UA.

## Steps

1. Set `ReactProbe = ReactProbeResolveUAOverChromeBoot`.

## Context

- Locked default: primary before connect is UA Firefox.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeResolveUAOverChromeBoot
	return nil
}
```
