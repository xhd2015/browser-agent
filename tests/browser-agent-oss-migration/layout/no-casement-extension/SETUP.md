# Scenario

**Feature**: Casement Chrome extension removed from repo root

```
# extension directory must not exist
Test Client -> stat RepoRoot/Chrome-Ext-Casement-Token
Test Client <- not found
```

## Preconditions

- `Chrome-Ext-Casement-Token/` is in the removed-artifacts list.

## Steps

1. Set `Leaf = no-casement-extension`.

## Context

- Absence check only at repo root.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Leaf = LeafNoCasementExtension
	return nil
}
```