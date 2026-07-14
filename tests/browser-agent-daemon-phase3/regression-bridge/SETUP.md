# Scenario

**Feature**: `browseragent.Run` backward compat — single session via registry

```
Test Client -> Run(ctx, Config{SessionID}) -> registry-backed control server
Test Client -> GET /v1/session?session=<live id> -> 200 snapshot
```

## Preconditions

- Mode `ModeRegression`.
- Uses full `Run` (not httptest-only handler).

## Steps

1. Set `Mode = ModeRegression`.
2. Leaf sets deterministic `SessionID`.

## Context

- Ensures `tests/browser-agent/` stays GREEN after phase 3 refactor.

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