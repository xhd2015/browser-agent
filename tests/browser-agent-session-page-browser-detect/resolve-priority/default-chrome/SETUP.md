# Scenario

**Feature**: resolveInstallBrowser defaults to chrome (P5)

```
resolveInstallBrowser
  # no forced, no EXT firefox, no Firefox UA, no snap/boot/path firefox
  -> "chrome"
```

## Preconditions

- Mode already `ModeReactSrc` from parent.

## Steps

1. Set `ReactProbe = ReactProbeResolveDefaultChrome`.

## Context

- Default remains chrome for unknown / Chrome operators.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeResolveDefaultChrome
	return nil
}
```
