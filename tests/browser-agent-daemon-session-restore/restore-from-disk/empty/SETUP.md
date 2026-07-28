# Scenario

**Feature**: Missing or empty sessions/ is a no-op restore

```
no sessions/ or empty sessions/ -> RestoreSessionsFromDisk -> List empty; err nil
```

## Preconditions

- No valid seed sessions (or only empty layout).
- BaseDir prepared.

## Steps

1. Ensure BaseDir and Addr.
2. Leaves set RestoreCase; Seeds empty.

## Context

- Missing sessions root is not an error (fresh BaseDir).

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	ensureBaseDir(t, req)
	ensureAddr(t, req)
	req.Seeds = nil
	return nil
}

// ensureEmptySessionsDir creates an empty sessions/ root (empty-sessions-dir leaf).
func ensureEmptySessionsDir(t *testing.T, req *Request) error {
	t.Helper()
	return os.MkdirAll(filepath.Join(req.BaseDir, "sessions"), 0o755)
}
```
