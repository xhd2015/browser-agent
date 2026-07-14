# Scenario

**Feature**: GET `/go?session=unknown` → 404

```
GET /go?session=does-not-exist -> 404
```

## Preconditions

- Unknown session id in query.

## Steps

1. Set `GoUnknownSession = true`.
2. Set `UnknownSessionID = "does-not-exist"`.

## Context

- Optional leaf from requirement; prevents serving arbitrary session page.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.GoUnknownSession = true
	req.UnknownSessionID = "does-not-exist"
	return nil
}
```