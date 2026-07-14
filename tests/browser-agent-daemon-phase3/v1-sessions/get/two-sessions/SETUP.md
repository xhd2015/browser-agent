# Scenario

**Feature**: GET `/v1/sessions` returns two sessions sorted ascending

```
POST sess-b, POST sess-a -> GET /v1/sessions -> [sess-a, sess-b]
```

## Preconditions

- Two distinct session ids created via POST before list.

## Steps

1. Set `PreCreateSessionIDs = []string{"sess-b"}`.
2. Set `SecondPreCreateID = "sess-a"` (created second; list must still sort).

## Context

- Sorted order is by `session_id` string ascending, not creation order.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PreCreateSessionIDs = []string{"sess-b"}
	req.SecondPreCreateID = "sess-a"
	return nil
}
```