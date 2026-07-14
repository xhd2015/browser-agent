# Scenario

**Feature**: /go install panel visible when not connected (requirement #6)

```
Control Server Session: no hello yet; extract already done
Test Client -> GET /go?session=<SessionSuffix>
HTML contains:
  - install panel marker
  - absolute path text (or path obtainable from page)
  - chrome://extensions text
  - Load unpacked / Developer guidance (preferred)
```

## Preconditions

- DoHello false.
- Product extracts on Run so path is known to the HTML renderer.

## Steps

1. Ensure `DoHello = false`.

## Context

- Path may be injected server-side into HTML; tests also accept a data attribute
  holding the path (e.g. data-extension-path="…").

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoHello = false
	return nil
}
```
