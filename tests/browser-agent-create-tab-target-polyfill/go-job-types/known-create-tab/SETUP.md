# Scenario

**Feature**: known create_tab true; prior six true; unknown false (G1)

```
IsKnownJobType("create_tab") -> true
IsKnownJobType("info"|"eval"|"run"|"logs"|"screenshot"|"cdp") -> true
IsKnownJobType(""|"foo"|"create-tab"|"CreateTab") -> false
```

## Preconditions

- Mode already go-job-types from parent.

## Steps

1. Ensure Mode stays ModeGoJobTypes (single leaf; Run uses fixed probe lists).

## Context

- Requirement G1. Hyphen form `create-tab` must remain unknown (CLI name ≠ job type).

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
