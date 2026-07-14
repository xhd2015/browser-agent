# Scenario

**Feature**: SessionDirExists returns true when session dir is present

```
MkdirAll({base}/sessions/{id}) -> SessionDirExists -> true
```

## Preconditions

- Session dir will be created before existence check.

## Steps

1. Set CreateSessionDir true.
2. SessionDirID `my-flow` (from parent default).

## Context

- Exists check reflects on-disk directory state.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CreateSessionDir = true
	return nil
}
```