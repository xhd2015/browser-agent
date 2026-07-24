# Scenario

**Feature**: extension connects with `browser-agent` support quickly

```
Fake Extension -> hello { version: "1.0.0", features: ["browser-agent"] }
Poll GET /v1/session?session=ID -> extension.connected=true, features includes browser-agent
  -> stderr "Extension connected", exit 0, elapsed < 5s, stdout has session output
```

## Preconditions

- `WaitOp = extension-connects-quickly`.
- Extension sends hello with version 1.0.0 and features including `browser-agent`.
- Wait timeout is short (5s) to avoid long test runs.

## Steps

1. Set `WaitOp = WaitOpExtensionConnectsQuickly`.
2. Set `SessionID = sess-new-wait-quick`.
3. Set `WaitExtensionTimeout = 5s` for fast feedback.
4. Set `ConnectExtension = true`, `SendHello = true`.
5. Set `HelloVersion = 1.0.0`, `HelloFeatures = ["browser-agent"]`.

## Context

- Extension connects within <5s (should be near-instant).
- Session output printed to stdout; extension connected message to stderr.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitOp = WaitOpExtensionConnectsQuickly
	req.SessionID = "sess-new-wait-quick"
	req.WaitExtensionTimeout = 5 * time.Second
	req.ConnectExtension = true
	req.SendHello = true
	req.HelloVersion = "1.0.0"
	req.HelloFeatures = []string{"browser-agent"}
	return nil
}
```
