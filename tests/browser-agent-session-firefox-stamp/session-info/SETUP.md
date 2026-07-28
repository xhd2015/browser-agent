# Scenario

**Feature**: session info Next steps follow firefox stamp when disconnected

```
SessionNew(Browser=firefox) -> session info
  -> Next steps: install-firefox-extension / about:debugging + firefox path
```

## Preconditions

- Mode = session-info.
- Extension remains disconnected (no WS connect in tests).

## Steps

1. Set Mode = ModeSessionInfo.

## Context

- Today FormatSessionInfo hardcodes install-chrome-extension; RED until firefox branch.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionInfo
	return nil
}
```
