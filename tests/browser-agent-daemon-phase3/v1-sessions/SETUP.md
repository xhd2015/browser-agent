# Scenario

**Feature**: HTTP session management — POST/GET `/v1/sessions`

```
Test Client -> httptest registry server
Test Client -> POST /v1/sessions | GET /v1/sessions
Registry Control Server -> Create | List snapshots
```

## Preconditions

- Registry-backed `NewRegistryControlHandler` (phase 3).

## Steps

1. Children set `Mode` to `ModeV1SessionsPost` or `ModeV1SessionsGet`.

## Context

- Distinct from package-only Phase 2 registry tests (wire JSON + status codes).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```