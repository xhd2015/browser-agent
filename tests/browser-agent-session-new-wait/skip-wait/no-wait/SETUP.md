# Scenario

**Feature**: new `--no-wait` flag skips wait entirely

```
NoWait=true
POST /v1/sessions -> session id
  -> no wait at all, exit 0, elapsed < 1s, stdout has session output
```

## Preconditions

- `SkipOp = no-wait`.
- `NoWait = true`.

## Steps

1. Set `SkipOp = SkipOpNoWait`.
2. Set `SessionID = sess-new-skip-no-wait`.
3. Set `NoWait = true`.

## Context

- The `--no-wait` flag (new `NoWait` field on `SessionNewConfig`) explicitly skips the wait.
- Returns immediately with normal session output.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SkipOp = SkipOpNoWait
	req.SessionID = "sess-new-skip-no-wait"
	req.NoWait = true
	return nil
}
```
