# Scenario

**Feature**: Run(SessionID) backward compat regression bridge

```
Run(ctx, Config{SessionID}) -> registry-backed single session serve
GET /v1/session?session=<id> -> 200
```

## Preconditions

- Mode `ModeRunCompat`.

## Steps

1. Set `Mode = ModeRunCompat`.
2. Leaf sets deterministic `SessionID`.

## Context

- Keeps `tests/browser-agent/` GREEN after daemon host split.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeRunCompat
	return nil
}
```