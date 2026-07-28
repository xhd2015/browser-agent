# Scenario

**Feature**: Valid session dirs with meta.json register successfully

```
valid id + meta.json -> RestoreSessionsFromDisk -> registered waiting_extension
```

## Preconditions

- Disk entries use valid session ids and well-formed meta.json.
- BaseDir prepared for seed writes.

## Steps

1. Ensure BaseDir and Addr for happy-path restores.
2. Leaves set RestoreCase and Seeds for cardinality.

## Context

- Outcome polarity: successful registration (not skip).

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
