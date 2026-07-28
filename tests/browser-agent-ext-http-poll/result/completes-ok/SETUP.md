# Scenario

**Feature**: result ok:true unblocks POST /v1/jobs after poll lease

```
Create + hello sess-ext-result-ok
POST /v1/jobs type=eval (bg)
POST /v1/ext/poll -> job_id
POST /v1/ext/result {session_id, job_id, ok:true, data:{source:"ext-poll"}}
  -> 200 {ok:true}
  -> /v1/jobs returns ok=true
```

## Preconditions

- ResultCase = completes-ok.
- ResultOK = true.
- ResultData set for observability.

## Steps

1. Full mini-pipeline in Run for ResultCompletesOK.

## Context

- Core success path for HTTP transport job completion.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ResultCase = ResultCompletesOK
	req.SessionID = "sess-ext-result-ok"
	req.ResultOK = true
	req.ResultData = map[string]any{"source": "ext-poll", "value": 2}
	req.JobHTTPType = "eval"
	req.JobHTTPParams = map[string]any{"expression": "1+1"}
	req.PollWaitMS = 2000
	return nil
}
```
