# Scenario

**Feature**: `POST` / `GET /v1/entries` JSON snapshot API (req #3, #4, #6)

```
Mock Extension -> POST /v1/entries {session_id, entries, count}
Test Client -> GET /v1/entries?session=<id>
Control Server -> {entries, count, updated_at} | 404
```

## Preconditions

- Final probe is `ProbeV1Entries`.
- Response is JSON (success or error).

## Steps

1. Set `Probe = ProbeV1Entries` (`"v1-entries"`).

## Context

- Split below on session identity (known vs missing), then snapshot sequence.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Probe = ProbeV1Entries
	return nil
}
```
