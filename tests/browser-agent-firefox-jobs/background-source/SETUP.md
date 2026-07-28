# Scenario

**Feature**: Firefox-Ext background.js source markers for Phase 2 core jobs

```
Read Firefox-Ext-Browser-Agent/public/background.js (build fallback)
  -> assert handleJob branches, job APIs, anti-stub contract
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
