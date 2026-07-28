# Scenario

**Feature**: result for unknown session returns 404

```
POST /v1/ext/result {session_id:"no-such-session-ext-result", job_id:"job-x", ok:true}
  -> 404
```

## Preconditions

- ResultCase = unknown-404.
- OmitCreate = true.
- ResultSessionID = no-such-session-ext-result.

## Steps

1. POST result with unknown session_id.

## Context

- Symmetric with hello/poll unknown guards.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ResultCase = ResultUnknown404
	req.OmitCreate = true
	req.SessionID = "sess-unused-result-404"
	req.ResultSessionID = "no-such-session-ext-result"
	req.ResultJobIDOverride = "job-x"
	req.ResultOK = true
	return nil
}
```
