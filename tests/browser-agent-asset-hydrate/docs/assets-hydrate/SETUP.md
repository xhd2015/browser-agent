# Scenario

**Feature**: hydrate operator doc covers cache, go install fallback, assets ensure

```
docs/assets-hydrate.md (preferred) and/or README / SKILL
  mentions:
    ~/.cache/browser-agent OR asset-cache
    go install OR incomplete embed
    assets ensure
```

## Preconditions

- DocsOp = assets-hydrate.

## Steps

1. Set `DocsOp = DocsOpAssetsHydrate`.

## Context

- Implementer may add `docs/assets-hydrate.md` or fold into README/SKILL.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DocsOp = DocsOpAssetsHydrate
	return nil
}
```
