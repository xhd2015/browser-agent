# Scenario

**Feature**: GET `/v1/session` without `session` query → 400

```
GET /v1/session (no ?session=) -> 400 bad request
```

## Preconditions

- Multi-session registry handler (session param required).

## Steps

1. Set `OmitSessionQuery = true`.

## Context

- Single-session `Run()` backward compat is covered under regression-bridge.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OmitSessionQuery = true
	return nil
}
```