# Scenario

**Feature**: Firefox content script registers session with background (like Chrome)

```
Session Page /go?session=S  (or data-session-id)
  -> contentScript reads session_id
  -> browser.runtime.sendMessage({type:"register", session_id:S, control_port, …})
  -> window.__BROWSER_AGENT_EXT__ retained
```

## Preconditions

- ExtSourceTarget = content-script-register.
- Source under Firefox-Ext-Browser-Agent (public preferred).

## Steps

1. Set `ExtSourceTarget = ExtSrcContentScriptRegister`.

## Context

- Accept `browser.runtime` or `chrome.runtime` sendMessage.
- Session id from URL query `session` and/or `data-session-id`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcContentScriptRegister
	return nil
}
```
