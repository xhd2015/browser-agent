# Scenario

**Feature**: React browser-trace product port 43759 (E2)

```
react/src/products/browser-trace.ts
  contains 43759 (+ browser-trace id)
```

## Preconditions

- ReactProductID = browser-trace.

## Steps

1. Set ReactProductID to browser-trace.

## Context

- Mirrors Go ProductBrowserTrace.ControlPort.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReactProductID = ReactProductTrace
	return nil
}
```
