# Scenario

**Feature**: background uses fetch (or HTTP POST client) for poll transport

```
fetch("http://127.0.0.1:{port}/v1/ext/poll", { method: "POST", body: … })
  # or equivalent POST client with hello/poll/result paths
```

## Preconditions

- BackgroundSourceTarget = http-poll-fetch.

## Steps

1. Set `BackgroundSourceTarget = BgSrcHttpPollFetch`.

## Context

- Prefer `fetch(`; explicit `method: "POST"` plus poll path language also OK.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcHttpPollFetch
	return nil
}
```
