# Scenario

**Feature**: serve-level --help text

```
HandleCLI serve --help -> serve-specific usage (not full session tree)
```

## Preconditions

- Mode `ModeHelp`.
- No daemon spawn.

## Steps

1. Set `Mode = ModeHelp`.

## Context

- Serve help must document new flags without requiring full `fullHelp` dump.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeHelp
	return nil
}
```