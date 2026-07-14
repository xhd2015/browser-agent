# Scenario

**Feature**: phase8 spawn-when-down parity with explicit --port

```
EnsureDaemon spawn on explicit --port N -> healthy + server.json
```

## Preconditions

- Mode `ModeRegression`.
- Leaf sets op-specific field.

## Steps

1. Set `Mode = ModeRegression`.

## Context

- See root DOCTEST for `Run` dispatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeRegression
	return nil
}
```
