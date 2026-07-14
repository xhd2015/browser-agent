# Scenario

**Feature**: GET `/v1/session?session=unknown` → 404

```
GET /v1/session?session=does-not-exist -> 404
```

## Preconditions

- No session with probe id.

## Steps

1. Set `ForceUnknownSession = true`.
2. Set `UnknownSessionID = "does-not-exist"`.

## Context

- Distinct from missing query param (400 leaf).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ForceUnknownSession = true
	req.UnknownSessionID = "does-not-exist"
	return nil
}
```