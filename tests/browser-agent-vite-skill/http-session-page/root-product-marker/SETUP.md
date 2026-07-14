# Scenario

**Feature**: GET / serves same SPA or redirects; product browser-agent present (A3)

```
GET /
  -> 200 HTML with product browser-agent
  OR 3xx to /go then 200 HTML with product browser-agent
```

## Preconditions

- HTTPProbe = root-product.

## Steps

1. Set HTTPProbe to root-product.
2. Harness may follow one redirect to obtain final HTML body.

## Context

- Current package may already serve `/` via handleGo; either path is fine.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HTTPProbe = HTTPProbeRoot
	return nil
}
```
