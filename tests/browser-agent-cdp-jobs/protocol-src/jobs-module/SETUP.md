# Scenario

**Feature**: jobs protocol module exists with six type constants (E1)

```
react/src/lib/protocol/jobs.ts|js
  tokens: info, eval, run, logs, screenshot, cdp
```

## Preconditions

- Mode already protocol-src from parent.

## Steps

1. Ensure Mode stays ModeProtocolSrc (single leaf under protocol-src).

## Context

- Requirement E1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Reaffirm mode for the single protocol-src leaf.
	req.Mode = ModeProtocolSrc
	return nil
}
```
