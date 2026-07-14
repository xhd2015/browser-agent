# Scenario

**Feature**: Go IsKnownJobType helper covers all six canonical types

```
Test Client -> IsKnownJobType(s) for each known type -> true
Test Client -> IsKnownJobType(unknown) -> false
```

## Preconditions

- Mode is go-job-types.
- Pure package call; no server.

## Steps

1. Set `Mode = ModeGoJobTypes`.

## Context

- Requirement F1. Export name `IsKnownJobType` preferred.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGoJobTypes
	return nil
}
```
