# Scenario

**Feature**: install markers present when extension not connected (A4)

```
GET /go (no WS hello)
  HTML has chrome://extensions OR Load unpacked OR install data marker
```

## Preconditions

- HTTPProbe = install-markers.
- No fake extension dialed.

## Steps

1. Set HTTPProbe to install-markers.

## Context

- Aligns with prior spa-embed/install-guideline-markers; keep markers GREEN.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HTTPProbe = HTTPProbeInstall
	return nil
}
```
