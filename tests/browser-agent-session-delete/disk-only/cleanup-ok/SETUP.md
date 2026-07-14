# Scenario

**Feature**: delete removes disk-only session directory

```
MkdirAll sessions/sess-disk-only-cleanup (no POST /v1/sessions)
HandleCLI session delete -> exit 0; dir gone
```

## Preconditions

- Session dir on disk only; not in registry.

## Steps

1. Set `DiskOnlySessionID = sess-disk-only-cleanup`.

## Context

- Daemon running for addr resolve from `server.json`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DiskOnlySessionID = "sess-disk-only-cleanup"
	return nil
}
```