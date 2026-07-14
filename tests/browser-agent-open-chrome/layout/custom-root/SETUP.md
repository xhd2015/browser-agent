# Scenario

**Feature**: custom --root override via LayoutFromRoot

```
LayoutFromRoot(temp/managed-chrome)
  -> data/ and extensions/ under that root
```

## Preconditions

- LayoutOp custom-root.
- ManagedRoot from root Setup temp dir.

## Steps

1. Set LayoutOp = LayoutOpCustomRoot.

## Context

- Mirrors CLI `--root` override semantics.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.LayoutOp = LayoutOpCustomRoot
	return nil
}
```
