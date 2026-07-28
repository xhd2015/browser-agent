# Scenario

**Feature**: background completes jobs via POST /v1/ext/result

```
handleJob(job)
  -> POST /v1/ext/result { session_id, job_id, ok, data?, error? }
```

## Preconditions

- BackgroundSourceTarget = result-endpoint.

## Steps

1. Set `BackgroundSourceTarget = BgSrcResultEndpoint`.

## Context

- Path token `/v1/ext/result` is required (parity with WS type=result).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcResultEndpoint
	return nil
}
```
