# Scenario

**Feature**: Dual React product modules on disk

```
ModuleRoot/react/src/products/browser-agent.ts  -> controlPort 43761
ModuleRoot/react/src/products/browser-trace.ts  -> controlPort 43759
```

## Preconditions

- ModeReactProducts.
- ModuleRoot from root Setup.
- ReactProductID set by leaf.
- No npm — existence + content probes only.

## Steps

1. Set Mode = ModeReactProducts.
2. Leave ReactProductID to leaf.

## Context

- Preferred path is `react/`; harness may also probe
  `project-api-capture-react/src/products/` as fallback.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeReactProducts
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	return nil
}
```
