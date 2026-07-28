# Scenario

**Feature**: Firefox-Ext background.js source markers for Phase 3 CDP matrix

```
Read Firefox-Ext-Browser-Agent/public/background.js (build fallback)
  -> assert handleJob case "cdp", method maps, unsupported message, anti-generic
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
