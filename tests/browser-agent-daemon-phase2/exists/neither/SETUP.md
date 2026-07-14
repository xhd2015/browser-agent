# Scenario

**Feature**: Exists false when absent from registry and disk

```
Exists(unknown) -> false
```

## Preconditions

- No pre-create; no disk seed.

## Steps

1. Set ExistsSessionID to `absent-id`.

## Context

- Negative path for Exists guard.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExistsSessionID = "absent-id"
	return nil
}
```