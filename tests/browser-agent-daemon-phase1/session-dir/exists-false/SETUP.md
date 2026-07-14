# Scenario

**Feature**: SessionDirExists returns false when session dir is absent

```
no mkdir -> SessionDirExists -> false
```

## Preconditions

- Session dir is not created.

## Steps

1. Set CreateSessionDir false (default).

## Context

- Absent directory must not report exists.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CreateSessionDir = false
	return nil
}
```