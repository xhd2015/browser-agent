# Scenario

**Feature**: service worker exposes status API for popup (message and/or storage)

```
popup (or future UI) needs session armed + debugger info
  -> chrome.runtime.sendMessage({ type: "status"|"getStatus"|"popupStatus" })
     and/or chrome.storage.session|local last-known write from SW
  -> payload includes session/WS state + debugger/attach set info
```

## Preconditions

- ExtSourceTarget = sw-status-api.

## Steps

1. Set `ExtSourceTarget = ExtSrcSWStatusAPI`.

## Context

- Current `background.js` `onMessage` handles only `register`; `sendSessionStatus`
  is **control-plane WS telemetry**, not a popup status API — **RED**.
- Implementer may choose message handler, storage push, or both; session armed +
  debugger info must be available.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcSWStatusAPI
	return nil
}
```
