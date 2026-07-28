# Scenario

**Feature**: Firefox-Ext background.js source markers for HTTP poll transport

```
Read Firefox-Ext-Browser-Agent/public/background.js (build fallback)
  -> assert /v1/ext/hello|poll|result, fetch, not WS-only, register→poll, prepare_reconnect
```

## Preconditions

- Mode = background-source.
- Sources under `Firefox-Ext-Browser-Agent` (public preferred).

## Steps

1. Set `Mode = ModeBackgroundSource`.
2. Leave `BackgroundSourceTarget` for each leaf.

## Context

- All leaves share the same background.js probe; Assert differs per target.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeBackgroundSource
	return nil
}
```
