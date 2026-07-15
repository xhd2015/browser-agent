# Scenario

**Feature**: MaxSeeds caps seed list after dedupe (0 keeps all)

```
testdata/max-seeds.md (5 unique URLs)
  MaxSeeds=0 -> 5 seeds
  MaxSeeds=2 -> 2 seeds
```

## Preconditions

- Mode is max-seeds.
- Fixture is max-seeds.md for both leaves.
- Leaf sets MaxSeeds.

## Steps

1. Set Mode to max-seeds.
2. Set Fixture to max-seeds.md.

## Context

- Cap applies after dedupe; Counts.AfterFilter reflects final length.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeMaxSeeds
	req.Fixture = "max-seeds.md"
	req.IncludeArchived = false
	return nil
}
```
