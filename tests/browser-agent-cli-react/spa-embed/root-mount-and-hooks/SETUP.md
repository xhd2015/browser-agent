# Scenario

**Feature**: SPA root mount + session poll hooks + product port (D1)

```
GET /go|/
  HTML has id="root" or data-browser-agent-root
  HTML references /v1/session
  HTML mentions 43761 and browser-agent
```

## Preconditions

- SPAProbe = root-hooks.

## Steps

1. Set SPAProbe to root-hooks.

## Context

- Session id may also appear in boot config (asserted softly via product markers).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SPAProbe = "root-hooks"
	return nil
}
```
