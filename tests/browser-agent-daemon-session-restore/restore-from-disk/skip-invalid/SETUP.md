# Scenario

**Feature**: Invalid disk entries are skipped without failing restore

```
bad dirname | missing meta | corrupt meta -> skip; continue; err nil
```

## Preconditions

- Seeds include one or more invalid layouts.
- BaseDir prepared for seed writes.

## Steps

1. Ensure BaseDir and Addr.
2. Leaves set RestoreCase and invalid Seeds.

## Context

- Best-effort: restore must not hard-fail on a single bad entry.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	ensureBaseDir(t, req)
	ensureAddr(t, req)
	return nil
}
```
