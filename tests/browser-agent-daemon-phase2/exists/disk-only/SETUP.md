# Scenario

**Feature**: Exists true when session dir on disk but not in registry

```
MkdirAll session dir (no Create) -> Exists(id) -> true
```

## Preconditions

- ExistsSeedDiskOnly true; no ExistsPreCreate.

## Steps

1. Set ExistsSessionID `disk-only`; ExistsSeedDiskOnly true.

## Context

- Simulates leftover dir after crash before registry restore.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExistsSessionID = "disk-only"
	req.ExistsSeedDiskOnly = true
	return nil
}
```