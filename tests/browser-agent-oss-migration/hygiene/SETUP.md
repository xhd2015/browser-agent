# Scenario

**Feature**: company words masked in published tree

```
# ripgrep for banned tokens (REQUIREMENT-DESIGN masking table originals)
Test Client -> rg -i <company-word-pattern> (excl .git, node_modules, dist)
Test Client <- no matches
```

## Preconditions

- Masking map applied per REQUIREMENT-DESIGN decision #5.
- `rg` on PATH.

## Steps

1. Set `Category = hygiene`.

## Context

- Exit code 1 from rg means no matches (success for this leaf).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Category = CategoryHygiene
	return nil
}
```