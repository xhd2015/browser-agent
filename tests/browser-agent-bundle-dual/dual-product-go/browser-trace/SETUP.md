# Scenario

**Feature**: ProductBrowserTrace defaults (C2)

```
ProductBrowserTrace
  id browser-trace
  controlPort 43759
  features include browser-trace
```

## Preconditions

- ProductProbe = browser-trace.

## Steps

1. Set ProductProbe to browser-trace.

## Context

- Capture-API / browser-trace product shell pairing.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ProductProbe = ProductProbeTrace
	return nil
}
```
