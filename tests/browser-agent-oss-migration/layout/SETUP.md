# Scenario

**Feature**: migrated repo layout — HAR subfolder present, Casement removed

```
# filesystem layout checks
Test Client -> stat paths under RepoRoot
Test Client <- har-viewer present; casement artifacts absent
```

## Preconditions

- HAR viewer moved to `har-viewer/` per locked decision #2.
- Casement extension/server stacks removed per decision #4.

## Steps

1. Set `Category = layout`.

## Context

- Read-only `os.Stat` / directory walk.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Category = CategoryLayout
	return nil
}
```