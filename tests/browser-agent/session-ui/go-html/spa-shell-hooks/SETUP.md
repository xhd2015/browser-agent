# Scenario

**Feature**: HTML embeds session SPA hooks (E3)

```
GET /go|/; body contains session id and /v1/session reference
  and a status/root marker for browser-agent session UI
```

## Preconditions

- No hello required for shell smoke.

## Steps

1. Leave DoWSHello false (shell is static/served).

## Context

- Markers may be data-browser-agent-* or generic session status ids.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Shell smoke: no extension hello; assert session + poller hooks only.
	req.DoWSHello = false
	req.Probe = ProbeGoHTML
	return nil
}
```

