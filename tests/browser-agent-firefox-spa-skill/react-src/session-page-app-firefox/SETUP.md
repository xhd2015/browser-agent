# Scenario

**Feature**: SessionPageApp wires Firefox into install UX (R2)

```
react/src/ui/SessionPageApp.tsx (+ InstallGuideline)
  detect firefox via boot / browsers / prop / product
  pass browser=firefox into InstallGuideline OR render firefox install steps
```

## Preconditions

- Mode already react-src from parent.

## Steps

1. Set `ReactProbe = ReactProbeSessionPageAppFirefox`.

## Context

- Combined SessionPageApp + InstallGuideline text is under test.
- Must not leave Chrome-only install when session browser is firefox.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeSessionPageAppFirefox
	return nil
}
```
