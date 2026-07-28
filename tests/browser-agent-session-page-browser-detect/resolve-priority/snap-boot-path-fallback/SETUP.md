# Scenario

**Feature**: snap / boot / path still yield firefox when higher signals absent (P4)

```
resolveInstallBrowser
  # no forced, no EXT firefox, no Firefox UA
  # snap.browser / snap.browsers / installPath browser-agent-firefox / boot
  -> "firefox"
```

## Preconditions

- Mode already `ModeReactSrc` from parent.
- Legacy signals remain as lower-priority fallback (not removed).

## Steps

1. Set `ReactProbe = ReactProbeResolveSnapBootPath`.

## Context

- Phase 2 will stamp boot correctly; client still honors snap/path when present.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeResolveSnapBootPath
	return nil
}
```
