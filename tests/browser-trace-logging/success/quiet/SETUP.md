# Scenario

**Feature**: Quiet mode suppresses info milestones

```
# Quiet: no progress noise on stderr; errors still allowed
browser-trace Quiet=true
Lifecycle Logger -/-> info milestones
browser-trace stdout -> "{sessionDir}\n"  # unchanged
```

## Preconditions

- `Quiet = true`.
- Success path still completes.

## Steps

1. Set `Quiet = true`.
2. Keep `Verbose = false` (Quiet wins if both ever set).

## Context

- Requirement scenario #2.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Quiet = true
	req.Verbose = false
	return nil
}
```
