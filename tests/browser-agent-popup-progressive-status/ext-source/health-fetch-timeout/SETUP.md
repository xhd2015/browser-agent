# Scenario

**Feature**: popup health/status fetch times out instead of hanging on checking…

```
popup.js health check
  -> fetch(/v1/health, { signal }) with AbortController
  -> abort after timeout (setTimeout / AbortSignal.timeout)
  -> catch -> unreachable/error (not infinite checking…)
```

## Preconditions

- ExtSourceTarget = health-fetch-timeout.

## Steps

1. Set `ExtSourceTarget = ExtSrcHealthFetchTimeout`.

## Context

- Current `popup.js` uses bare `fetch("http://127.0.0.1:43761/v1/health")` with no
  abort/timeout — **RED** until implementer adds AbortController (or equivalent).
- Promise.race with a timeout rejection is also acceptable.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcHealthFetchTimeout
	return nil
}
```
