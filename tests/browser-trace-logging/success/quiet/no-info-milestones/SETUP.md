# Scenario

**Feature**: quiet success — stderr free of info progress; stdout path only

```
# Quiet success
Mock Extension completes
Lifecycle Logger Quiet -> stderr has no listen/ready/recording progress lines
stdout -> "{sessionDir}\n"
# Log file also suppressed under Quiet
```

## Preconditions

- Inherits Quiet from parent.
- Default log-file policy; Quiet implies no info log file content (file absent
  or empty of milestones — assert no info file when Quiet).

## Steps

1. Confirm Quiet remains true.
2. Run success path.

## Context

- Requirement #2: stderr has no info milestones (empty or only unexpected warnings).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Quiet = true
	// Explicit: quiet should not write progress log file either.
	req.NoLogFile = false
	return nil
}
```
