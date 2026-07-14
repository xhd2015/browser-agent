# Scenario

**Feature**: Agent and Trace control ports differ (C3)

```
ProductBrowserAgent.ControlPort != ProductBrowserTrace.ControlPort
  43761 vs 43759
```

## Preconditions

- ProductProbe = ports-differ.
- Run loads both product configs.

## Steps

1. Set ProductProbe to ports-differ.

## Context

- Explicit non-collision leaf beyond per-product checks.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ProductProbe = ProductProbePortsDiffer
	return nil
}
```
