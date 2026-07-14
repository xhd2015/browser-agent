# Scenario

**Feature**: SessionRegistry Exists — registry or disk

```
Test Client -> Exists(id) -> true if in map OR SessionDirExists
```

## Preconditions

- Mode is exists.
- Leaf configures pre-create or disk-only seed.

## Steps

1. Set Mode to exists.
2. Ensure BaseDir and Addr.

## Context

- Crash recovery: disk dir without registry entry still counts as exists.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExists
	ensureBaseDir(t, req)
	ensureAddr(t, req)
	return nil
}
```