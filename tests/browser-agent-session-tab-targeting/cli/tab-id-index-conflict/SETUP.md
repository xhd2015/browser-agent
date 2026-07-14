# Scenario

**Feature**: --tab-id and --tab-index are mutually exclusive

```
HandleCLI session eval --tab-id 111 --tab-index 2 -> error before job POST
```

## Preconditions

- CLIOp = tab-id-index-conflict.
- No daemon required.

## Steps

1. Set `CLIOp = CLIOpTabIDIndexConflict`.

## Context

- Spec error: `cannot use both --tab-id and --tab-index`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpTabIDIndexConflict
	return nil
}
```