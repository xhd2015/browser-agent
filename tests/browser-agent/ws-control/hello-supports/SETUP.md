# Scenario

**Feature**: WS hello with browser-agent feature → supports true (D1)

```
Fake Extension -> WS hello {version≥floor, features:[browser-agent]}
Test Client -> GET /v1/session
  -> extension.connected=true, supports_browser_agent=true
```

## Preconditions

- HelloVersion 1.0.0; features include browser-agent.
- WSAction hello-supports.

## Steps

1. Set WSAction to hello-supports.

## Context

- Floor version is product-defined; harness uses 1.0.0 as ≥ floor.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WSAction = WSActionHelloSupports
	req.HelloVersion = "1.0.0"
	req.HelloFeatures = []string{"browser-agent"}
	return nil
}
```
