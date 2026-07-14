# Scenario

**Feature**: POST /v1/jobs against the live session id

```
Control Server session = req.SessionID (known)
Test Client -> POST /v1/jobs session_id=<live>
  -> 200 + JobResult (complete or timeout) — not 404
```

## Preconditions

- ForceUnknownSession is false.
- Live session from root Setup.

## Steps

1. Ensure `ForceUnknownSession = false`.
2. Children toggle fake extension / timeout.

## Context

- C1 success vs C2 timeout are mutually exclusive children.

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
