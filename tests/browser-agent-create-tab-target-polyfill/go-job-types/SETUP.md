# Scenario

**Feature**: Go IsKnownJobType accepts create_tab additively

```
Test Client -> IsKnownJobType("create_tab") -> true
Test Client -> IsKnownJobType(prior six) -> true
Test Client -> IsKnownJobType(unknown) -> false
# never assert exclusive set size
```

## Preconditions

- Mode is go-job-types.
- Pure package call; no server.

## Steps

1. Set `Mode = ModeGoJobTypes`.

## Context

- Requirement G1. Additive only so sealed cdp-jobs stays GREEN.

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
