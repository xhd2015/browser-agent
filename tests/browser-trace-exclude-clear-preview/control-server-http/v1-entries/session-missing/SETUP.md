# Scenario

**Feature**: GET /v1/entries unknown session → 404 JSON (req #6)

```
Test Client -> GET /v1/entries?session=does-not-exist
Control Server -> HTTP 404 JSON not-found
```

## Preconditions

- Live server is up with a real session, but probe uses a different id.
- No POST staging required for the GET-missing path.

## Steps

1. Set `ForceUnknownSession = true`.
2. Leave default unknown id (`does-not-exist`) unless overridden.
3. Do not stage posts (`DoStagePost = false`).

## Context

- Sibling of `session-known/`.
- POST-to-unknown is product-defensible as 404 too; this leaf asserts **GET**.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ForceUnknownSession = true
	req.SessionIDForProbe = "does-not-exist"
	req.DoStagePost = false
	req.DoClearAfterStage = false
	return nil
}
```
