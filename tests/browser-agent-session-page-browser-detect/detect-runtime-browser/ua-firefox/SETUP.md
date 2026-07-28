# Scenario

**Feature**: detectRuntimeBrowser maps Firefox UA to firefox (D1)

```
detectRuntimeBrowser
  # userAgent contains Firefox/
  -> "firefox"
```

## Preconditions

- Mode already `ModeReactSrc` from parent.

## Steps

1. Set `ReactProbe = ReactProbeDetectUAFirefox`.

## Context

- Source must name `detectRuntimeBrowser` and use a Firefox UA token.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeDetectUAFirefox
	return nil
}
```
