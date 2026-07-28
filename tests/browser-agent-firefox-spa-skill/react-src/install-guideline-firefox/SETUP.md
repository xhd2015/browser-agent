# Scenario

**Feature**: InstallGuideline Firefox install steps (R1)

```
react/src/ui/InstallGuideline.tsx
  when browser is firefox (prop / dual path):
    about:debugging (+ optional #/runtime/this-firefox)
    Load Temporary Add-on / temporary add-on
    browser-agent-firefox path segment
```

## Preconditions

- Mode already react-src from parent.

## Steps

1. Set `ReactProbe = ReactProbeInstallGuidelineFirefox`.

## Context

- Implementer may use `browser?: "chrome" | "firefox"` prop or conditional branches.
- All three markers required for GREEN.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeInstallGuidelineFirefox
	return nil
}
```
