# Scenario

**Feature**: no version check

```
serve --kill-existing -> warn connected + RemoveAll orphan dirs
```

## Preconditions

- `KillExistingOp = KillExistingOpNoVersionCheck`.

## Steps

1. Set `KillExistingOp = KillExistingOpNoVersionCheck`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.KillExistingOp = KillExistingOpNoVersionCheck
	return nil
}
```
