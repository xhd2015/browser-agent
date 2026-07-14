# Scenario

**Feature**: mutually exclusive serve mode flags

```
HandleCLI serve with >1 of --stop, --status, --kill-existing -> fatal error exit 1
```

## Preconditions

- Mode `ModeModeConflict`.
- Parse must fail before daemon operations.

## Steps

1. Set `Mode = ModeModeConflict`.

## Context

- Error should mention the three mode flags are mutually exclusive.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeModeConflict
	return nil
}
```