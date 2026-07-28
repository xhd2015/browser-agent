# Scenario

**Feature**: Chrome shell background.js implements prepare_reconnect handling

```
Chrome-Ext-Browser-Agent/public/background.js
  -> handle type prepare_reconnect
  -> read delay_ms; schedule WS close; aggressive reconnect
```

## Preconditions

- ExtSourceTarget = chrome-background.
- Candidates: public/, root, src/, build/, embedded/extension/.

## Steps

1. Set ExtSourceTarget chrome-background.

## Context

- Existing reconnect backoff is insufficient alone — must react to control-plane prepare_reconnect.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcChromeBackground
	return nil
}
```
