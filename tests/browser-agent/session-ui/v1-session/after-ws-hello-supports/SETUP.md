# Scenario

**Feature**: after WS hello supports → connected + supports_browser_agent (E2)

```
Fake Extension WS hello feature browser-agent
GET /v1/session -> connected=true, supports_browser_agent=true
```

## Preconditions

- DoWSHello true with supporting hello payload.

## Steps

1. Set DoWSHello true.
2. HelloVersion 1.0.0; features browser-agent.

## Context

- Mirrors D1 snapshot path under session-ui surface.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoWSHello = true
	req.HelloVersion = "1.0.0"
	req.HelloFeatures = []string{"browser-agent"}
	return nil
}
```
