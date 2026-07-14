# Scenario

**Feature**: ProductBrowserTrace dual export (C2)

```
ProductBrowserTrace
  id browser-trace
  controlPort 43759
```

## Preconditions

- ProductID = browser-trace.
- Dual export required by this tree (shared ProductConfig design).

## Steps

1. Set ProductID to browser-trace.

## Context

- Keeps agent/trace parameterization data-driven for InstallGuideline.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ProductID = "browser-trace"
	return nil
}
```
