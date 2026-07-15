# Scenario

**Feature**: skip or include URLs under historical/archived/deprecated headings

```
testdata/historical.md
  headings Historical | Archived | Deprecated
  -> default: skip section body links
  -> IncludeArchived: keep them
```

## Preconditions

- Mode is historical.
- Fixture is historical.md for both leaves.
- Leaf sets IncludeArchived true or false.

## Steps

1. Set Mode to historical.
2. Set Fixture to historical.md.

## Context

- Skip until next same-or-higher heading level.
- Active section links always included.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeHistorical
	req.Fixture = "historical.md"
	req.MaxSeeds = 0
	return nil
}
```
