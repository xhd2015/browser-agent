# Scenario

**Feature**: /preview against the live session id

```
GET /preview?session=<live> -> 200 HTML
```

## Preconditions

- Session id is the live `SessionSuffix`.
- Children stage entries (or clear) before GET.

## Steps

1. Ensure `ForceUnknownSession = false`.

## Context

- Sibling of `session-missing/`.

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
