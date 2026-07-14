# Scenario

**Feature**: SessionRegistry Get — lookup live session

```
Test Client -> Get(id) -> (*session, bool)
```

## Preconditions

- Mode is get.
- Leaf sets GetSessionID; found leaf pre-creates via SessionID.

## Steps

1. Set Mode to get.
2. Ensure BaseDir and Addr.

## Context

- Get only sees registry entries, not disk-only dirs.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGet
	ensureBaseDir(t, req)
	ensureAddr(t, req)
	return nil
}
```