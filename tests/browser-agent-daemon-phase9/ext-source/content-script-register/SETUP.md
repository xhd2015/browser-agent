# Scenario

**Feature**: content script registers session tab with background (P3)

```
Session Page /go?session=S
  -> contentScript reads session id from URL
  -> chrome.runtime.sendMessage({type:"register", session_id:S, tabId, windowId})
```

## Preconditions

- ExtSourceTarget = content-script-register.

## Steps

1. Set ExtSourceTarget content-script-register.

## Context

- Session page must trigger background per-session WS connect for that tab.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcContentScriptRegister
	return nil
}
```