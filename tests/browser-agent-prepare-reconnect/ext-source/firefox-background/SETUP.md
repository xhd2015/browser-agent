# Scenario

**Feature**: Firefox shell background.js implements prepare_reconnect handling

```
Firefox-Ext-Browser-Agent/public/background.js
  -> handle type prepare_reconnect
  -> read delay_ms; schedule WS close; aggressive reconnect
```

## Preconditions

- ExtSourceTarget = firefox-background.
- Candidates: public/, root, src/, build/, embedded/extension-firefox/.

## Steps

1. Set ExtSourceTarget firefox-background.

## Context

- Same control-plane contract as Chrome; Firefox must not lag on upgrade reconnect.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcFirefoxBackground
	return nil
}
```
