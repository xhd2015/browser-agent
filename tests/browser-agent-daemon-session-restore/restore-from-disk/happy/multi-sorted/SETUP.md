# Scenario

**Feature**: Multiple restored sessions appear sorted by session id

```
seed sess-bbb + sess-aaa -> RestoreSessionsFromDisk
  -> List ids [sess-aaa, sess-bbb]
```

## Preconditions

- RestoreCase = multi-sorted.
- Two seeds created out of sort order.

## Steps

1. Set RestoreCase and Seeds (`sess-bbb` then `sess-aaa`).

## Context

- List order is sorted by id, not seed order.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.RestoreCase = RestoreCaseMultiSorted
	req.Seeds = []SeedSession{
		{ID: "sess-bbb"},
		{ID: "sess-aaa"},
	}
	return nil
}
```
