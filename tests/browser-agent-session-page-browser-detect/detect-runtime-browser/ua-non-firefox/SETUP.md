# Scenario

**Feature**: detectRuntimeBrowser maps non-Firefox UA to chrome (D2)

```
detectRuntimeBrowser
  # Chrome / Safari / empty / missing navigator
  -> "chrome"
```

## Preconditions

- Mode already `ModeReactSrc` from parent.

## Steps

1. Set `ReactProbe = ReactProbeDetectUANonFirefox`.

## Context

- Default branch must be chrome (not firefox-only helper).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeDetectUANonFirefox
	return nil
}
```
