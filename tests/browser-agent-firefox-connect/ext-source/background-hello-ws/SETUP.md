# Scenario

**Feature**: Firefox background opens per-session WS and sends hello with browser_product firefox

```
Background on register
  -> new WebSocket(…/v1/ws?session=S)
  -> hello { version, features, browser_product: "firefox", bundle_md5? }
```

## Preconditions

- ExtSourceTarget = background-hello-ws.
- Source under Firefox-Ext-Browser-Agent (public preferred).

## Steps

1. Set `ExtSourceTarget = ExtSrcBackgroundHelloWS`.

## Context

- Features should include browser-agent (asserted softly via hello + features language when present).
- `browser_product` key or literal firefox identity required.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcBackgroundHelloWS
	return nil
}
```
