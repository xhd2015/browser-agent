# Scenario

**Feature**: SessionNew stdout Extension block + Chrome 137 note

```
SessionNew -> stdout: path + install-chrome-extension + Chrome 137 cannot auto-load
```

## Preconditions

- Pretty stdout after successful create.

## Steps

1. Set `SessionNewOp = stdout-extension-block`.

## Context

- Uses v2 output markers (substring assertions).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpStdoutExtensionBlock
	return nil
}
```