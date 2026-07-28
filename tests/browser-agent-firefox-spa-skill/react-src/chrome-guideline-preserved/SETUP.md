# Scenario

**Feature**: InstallGuideline still documents Chrome path (R3)

```
react/src/ui/InstallGuideline.tsx
  chrome://extensions remains for Chrome / default browser path
```

## Preconditions

- Mode already react-src from parent.
- Out of scope: removing Chrome default install path.

## Steps

1. Set `ReactProbe = ReactProbeChromeGuidelinePreserved`.

## Context

- Expect **GREEN** already (current component is Chrome-primary).
- Implementer must not replace Chrome-only copy with Firefox-only copy.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeChromeGuidelinePreserved
	return nil
}
```
