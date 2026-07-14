# Scenario

**Feature**: List returns sessions sorted by id ascending

```
Create(sess-bbb) then Create(sess-aaa) -> List -> [sess-aaa, sess-bbb]
```

## Preconditions

- Two session ids out of sort order for create sequence.

## Steps

1. Set ListSessionIDs to `sess-bbb`, `sess-aaa` (creation order).

## Context

- List order must be sorted by id, not creation order.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListSessionIDs = []string{"sess-bbb", "sess-aaa"}
	return nil
}
```