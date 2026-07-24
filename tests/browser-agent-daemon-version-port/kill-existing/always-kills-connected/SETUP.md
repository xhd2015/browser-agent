# Scenario

**Feature**: always kills connected

```
serve --kill-existing -> warn connected + RemoveAll orphan dirs
```

## Preconditions

- `KillExistingOp = KillExistingOpAlwaysKillsConnected`.

## Steps

1. Set `KillExistingOp = KillExistingOpAlwaysKillsConnected`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.KillExistingOp = KillExistingOpAlwaysKillsConnected
	return nil
}
```
