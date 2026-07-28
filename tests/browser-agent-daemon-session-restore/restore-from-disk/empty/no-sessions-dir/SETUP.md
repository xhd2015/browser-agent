# Scenario

**Feature**: Absent sessions/ directory restores empty with nil error

```
BaseDir without sessions/ -> RestoreSessionsFromDisk -> List empty; err nil
```

## Preconditions

- RestoreCase = no-sessions-dir.
- Seeds empty; do not create sessions/.

## Steps

1. Set RestoreCase; leave Seeds empty.

## Context

- Fresh BaseDir has no sessions root — not an error.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.RestoreCase = RestoreCaseNoSessionsDir
	req.Seeds = nil
	return nil
}
```
