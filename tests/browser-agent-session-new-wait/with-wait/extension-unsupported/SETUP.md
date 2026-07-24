# Scenario

**Feature**: extension connects but does NOT support browser-agent

```
Fake Extension -> hello { version: "0.9.0", features: ["something-else"] }
Poll GET /v1/session?session=ID -> connected but features does NOT include browser-agent
  -> error "does not support browser-agent", exit 1
```

## Preconditions

- `WaitOp = extension-unsupported`.
- Extension sends hello with version < 1.0.0 OR features missing `browser-agent`.
- Test uses version 0.9.0 and empty/irrelevant features.

## Steps

1. Set `WaitOp = WaitOpExtensionUnsupported`.
2. Set `SessionID = sess-new-wait-unsupported`.
3. Set `WaitExtensionTimeout = 5s` for fast feedback.
4. Set `ConnectExtension = true`, `SendHello = true`.
5. Set `HelloVersion = 0.9.0`, `HelloFeatures = []`.

## Context

- Even though extension is connected, it does not support browser-agent.
- Error message: "Error: extension does not support browser-agent".
- Exit code must be non-zero.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitOp = WaitOpExtensionUnsupported
	req.SessionID = "sess-new-wait-unsupported"
	req.WaitExtensionTimeout = 5 * time.Second
	req.ConnectExtension = true
	req.SendHello = true
	req.HelloVersion = "0.9.0"
	req.HelloFeatures = []string{}
	return nil
}
```
