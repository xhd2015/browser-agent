# Scenario

**Feature**: unsupported CDP methods return clear firefox-specific not-supported errors

```
handleCdpJob / cdp branch
  method not in matrix (e.g. Network.enable, DOM.getDocument)
    -> throw / sendJobResult error
    message contains "not supported" (or "unsupported") AND "firefox"
```

## Preconditions

- BackgroundSourceTarget = unsupported-method-message.

## Steps

1. Set `BackgroundSourceTarget = BgSrcUnsupportedMethodMessage`.

## Context

- Product-owned wording required; case-insensitive match on both tokens.
- Example shape: `"CDP method X is not supported on firefox"`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcUnsupportedMethodMessage
	return nil
}
```
