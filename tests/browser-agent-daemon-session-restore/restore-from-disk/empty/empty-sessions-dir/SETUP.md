# Scenario

**Feature**: Empty sessions/ directory restores empty with nil error

```
mkdir sessions/ (empty) -> RestoreSessionsFromDisk -> List empty; err nil
```

## Preconditions

- RestoreCase = empty-sessions-dir.
- Empty `sessions/` created under BaseDir.

## Steps

1. Set RestoreCase.
2. Create empty sessions/ directory.

## Context

- Distinguishes empty root from missing root.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.RestoreCase = RestoreCaseEmptySessionsDir
	req.Seeds = nil
	return ensureEmptySessionsDir(t, req)
}
```
