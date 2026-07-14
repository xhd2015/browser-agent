# Scenario

**Feature**: react product file for browser-agent (G1)

```
react/src/products/browser-agent.ts(.tsx|.js) exists
  content contains 43761 and browser-agent
```

## Preconditions

- ReactProbe = products-browser-agent.

## Steps

1. Set ReactProbeProducts.

## Context

- TS mirror of ProductConfig.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeProducts
	return nil
}
```
