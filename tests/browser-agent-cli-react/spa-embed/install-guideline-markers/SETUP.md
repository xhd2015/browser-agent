# Scenario

**Feature**: install guideline markers when extension not connected (D2)

```
GET /go|/ without WS hello
  HTML shows install guidance:
    chrome://extensions and/or Load unpacked
```

## Preconditions

- SPAProbe = install-guideline.
- No extension WS connection.

## Steps

1. Set SPAProbe to install-guideline.

## Context

- chrome:// as text (copyable); links from http origin are blocked by browsers.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SPAProbe = "install-guideline"
	return nil
}
```
