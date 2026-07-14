# Scenario

**Feature**: /v1/entries against the live session id

```
# probe and POST use SessionSuffix (live session)
POST /v1/entries session_id=<live>
GET  /v1/entries?session=<live>
```

## Preconditions

- Session id is the live `SessionSuffix` (not ForceUnknownSession).
- Children stage entry snapshots then GET.

## Steps

1. Ensure `ForceUnknownSession = false`.
2. Leave stage/clear flags to leaves.

## Context

- Sibling of `session-missing/` under MECE on session identity.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ForceUnknownSession = false
	return nil
}
```
