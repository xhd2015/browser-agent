# Scenario

**Feature**: wait is skipped entirely — no extension polling

```
POST /v1/sessions -> session id
NoOpenChrome=true or NoWait=true
  -> no wait, exit 0, elapsed < 1s, stdout has session output
```

## Preconditions

- Mode is `skip-wait`.
- Either `NoOpenChrome` or `NoWait` is true.
- Daemon is running, session created.

## Steps

1. Set `Mode = ModeSkipWait`.
2. Leaves set `SkipOp` to select the skip reason.

## Context

- No extension polling at all.
- Session output printed directly to stdout.
- Returns immediately (under 1s).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeSkipWait
	return nil
}
```
