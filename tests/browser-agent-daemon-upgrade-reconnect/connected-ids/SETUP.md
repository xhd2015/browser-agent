# Scenario

**Feature**: Pure ConnectedIDsFromSnapshots builds waitList from session connectivity views

```
ConnectedIDsFromSnapshots([{session_id, connected}, ...])
  -> [ids where connected=true], input order
```

## Preconditions

- Mode = `connected-ids`.
- Leaf sets `ConnectedIDsCase` and `Snapshots`.

## Steps

1. Set `Mode = ModeConnectedIDs`.

## Context

- No HTTP; pure function only.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeConnectedIDs
	return nil
}
```
