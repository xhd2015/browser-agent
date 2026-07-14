# Scenario

**Feature**: session info after fake WS hello merges info job tabs (B2)

```
serve + fake extension hello
fake extension auto-completes job type=info with tabs[]
HandleCLI session info --session-id --addr
  -> extension.connected true
  -> stdout includes browser/tabs (or nested data with tabs)
  -> trailing \n
```

## Preconditions

- Live serve; FakeExtension true with info job tabs payload.
- SessionInfoKind = connected-with-info-tabs.

## Steps

1. Set SessionInfoKind connected-with-info-tabs.
2. Provide InfoJobTabs fixture (two http pages).
3. Leave CLIArgs empty for harness injection.
4. MaxDispatchWait generous for job wait.

## Context

- Requirement B2. Implementer must POST /v1/jobs type=info when connected.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionInfoKind = SessionInfoConnectedWithInfoTabs
	req.FakeExtension = true
	req.CLIArgs = nil
	req.MaxDispatchWait = 12 * time.Second
	req.InfoVersion = "1.0.0"
	req.InfoJobTabs = []map[string]any{
		{"id": float64(11), "url": "https://example.com/", "title": "Example"},
		{"id": float64(22), "url": "https://shop.example/cart", "title": "Cart"},
	}
	return nil
}
```
