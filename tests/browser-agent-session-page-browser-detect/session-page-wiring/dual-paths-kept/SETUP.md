# Scenario

**Feature**: keep InstallGuideline dual chrome + firefox paths (W3)

```
InstallGuideline + SessionPageApp
  firefox: about:debugging
  chrome: chrome://extensions
```

## Preconditions

- Mode already `ModeReactSrc` from parent.
- Phase 1 must not collapse to a single-browser install panel.

## Steps

1. Set `ReactProbe = ReactProbeWiringDualPaths`.

## Context

- Regression guard while priority helpers change.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeWiringDualPaths
	return nil
}
```
