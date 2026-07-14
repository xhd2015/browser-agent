# Scenario

**Feature**: GET /go returns session id + boot/product markers + poll hooks (A2)

```
GET /go?session=<id>
  body: session id
  body: browser-agent + 43761 + /v1/session
  body: boot script or data-* / __BROWSER_AGENT
  body: root mount
```

## Preconditions

- HTTPProbe = go-boot.

## Steps

1. Set HTTPProbe to go-boot.
2. Server injects live SessionID into HTML/boot.

## Context

- Stronger than prior spa-embed: requires live session id in body.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HTTPProbe = HTTPProbeGoBoot
	return nil
}
```
