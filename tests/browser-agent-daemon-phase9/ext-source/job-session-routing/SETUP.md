# Scenario

**Feature**: background routes jobs to tab by session_id (P4)

```
Background on job (payload.session_id = S)
  -> pickTargetTabId(S) or equivalent
  -> prefer registered tabId/windowId
  -> fallback URL match /go?session=S
```

## Preconditions

- ExtSourceTarget = job-session-routing.

## Steps

1. Set ExtSourceTarget job-session-routing.

## Context

- Server job payload already carries session_id; extension must honor it.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcJobSessionRouting
	return nil
}
```