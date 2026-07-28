# Scenario

**Feature**: handleJob dispatches a dedicated branch for job type cdp

```
handleJob(msg)
  switch jobType
    case "cdp"
      -> handleCdpJob (or equivalent method matrix)
      # not only default: unknown job type
```

## Preconditions

- BackgroundSourceTarget = handle-job-cdp-case.

## Steps

1. Set `BackgroundSourceTarget = BgSrcHandleJobCdpCase`.

## Context

- Quoted / `case` / `===` forms preferred via `jobTypeTokenPresent`.
- Named `handleCdpJob` is also accepted as dedicated surface.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcHandleJobCdpCase
	return nil
}
```
