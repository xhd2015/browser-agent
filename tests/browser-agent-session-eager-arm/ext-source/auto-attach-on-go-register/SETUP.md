# Scenario

**Feature**: `/go` register connects WebSocket only (no debugger attach)

```
handleRegisterMessage / maybeRegisterGoTab
  -> bind session tabId + windowId
  -> connectSession(...)
  -> NO chrome.debugger.attach on the control page
```

## Preconditions

- ExtSourceTarget = auto-attach-on-go-register.

## Steps

1. Set `ExtSourceTarget = ExtSrcAutoAttachOnGoRegister`.

## Context

- Product decision (popup UX): attaching debugger on `/go` register freezes Chrome and
  delays the toolbar `default_popup` for tens of seconds.
- Content tabs still eager-attach on `create_tab` / navigate (sibling leaves).
- Jobs attach on demand via `withDebuggerForSession`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcAutoAttachOnGoRegister
	return nil
}
```
