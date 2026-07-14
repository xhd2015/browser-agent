# Scenario

**Feature**: --color and --no-color cannot be used together

```
HandleCLI serve --color --no-color --stop -> fatal error exit 1
```

## Preconditions

- Empty base-dir.
- Both color flags on same invocation.

## Steps

1. Set `ColorOp = ColorOpColorConflict`.

## Context

- cli-color mutual exclusion for conflicting color flags.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ColorOp = ColorOpColorConflict
	return nil
}
```