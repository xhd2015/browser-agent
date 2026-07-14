# Scenario

**Feature**: GET `/go?session=<id>` — session page HTML

```
Test Client -> GET /go?session=<id>
Registry Control Server -> SPA HTML + session warning banner
```

## Preconditions

- Mode `ModeGoHTML`.

## Steps

1. Set `Mode = ModeGoHTML`.
2. Leaves probe known session HTML or unknown 404.

## Context

- Warning marker: `data-browser-agent-session-warning`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGoHTML
	return nil
}
```