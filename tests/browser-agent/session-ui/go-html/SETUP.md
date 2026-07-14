# Scenario

**Feature**: GET /go or / session SPA HTML for browser-agent product

```
Test Client -> GET /go?session=<id> (fallback /)
HTML -> session id hooks, /v1/session poll, product port/name markers
```

## Preconditions

- Probe is go-html.
- No DOM automation.

## Steps

1. Set `Probe = ProbeGoHTML`.
2. Children assert shell hooks vs product port strings.

## Context

- E3 SPA shell; E4 product parameterization (43761).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Probe = ProbeGoHTML
	req.DoWSHello = false
	return nil
}
```
