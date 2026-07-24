# Scenario

**Feature**: `--no-open-chrome` skips wait entirely (backward compat)

```
NoOpenChrome=true
POST /v1/sessions -> session id
  -> no wait at all, exit 0, elapsed < 1s, stdout has session output
```

## Preconditions

- `SkipOp = no-open-chrome`.
- `NoOpenChrome = true`.

## Steps

1. Set `SkipOp = SkipOpNoOpenChrome`.
2. Set `SessionID = sess-new-skip-no-open-chrome`.
3. Set `NoOpenChrome = true`.

## Context

- When Chrome isn't opened, there is nothing to wait for.
- Returns immediately with normal session output.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SkipOp = SkipOpNoOpenChrome
	req.SessionID = "sess-new-skip-no-open-chrome"
	req.NoOpenChrome = true
	return nil
}
```
