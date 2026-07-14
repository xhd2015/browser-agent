# Scenario

**Feature**: serve --help documents stop and color flags

```
HandleCLI serve --help -> stdout contains --stop, --color, --no-color
```

## Preconditions

- Default args `serve --help`.

## Steps

1. Leaf uses default help args from Run.

## Context

- Serve-specific help contract for less-flags + cli-color migration.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```