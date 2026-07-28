# Scenario

**Feature**: helpers named detectRuntimeBrowser + resolveInstallBrowser (W1)

```
SessionPageApp / ui helpers
  detectRuntimeBrowser  (named)
  resolveInstallBrowser (named)
```

## Preconditions

- Mode already `ModeReactSrc` from parent.

## Steps

1. Set `ReactProbe = ReactProbeWiringHelpersNamed`.

## Context

- Readable source contracts for implementer and future unit extraction.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeWiringHelpersNamed
	return nil
}
```
