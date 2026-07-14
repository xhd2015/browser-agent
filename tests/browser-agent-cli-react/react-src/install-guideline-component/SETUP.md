# Scenario

**Feature**: InstallGuideline React component file (G3)

```
react/src/ui/InstallGuideline.tsx (or .ts / .jsx / components/)
```

## Preconditions

- ReactProbe = install-guideline.

## Steps

1. Set ReactProbeInstallGuideline.

## Context

- Component is parameterized by ProductConfig in product code (not asserted here).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeInstallGuideline
	return nil
}
```
