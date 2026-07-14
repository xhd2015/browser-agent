# Scenario

**Feature**: default info + session log file enabled (product default)

```
# Log file mirrors info+ lines
Lifecycle Logger -> stderr (info milestones)
Lifecycle Logger -> {sessionDir}/browser-trace.log
NoLogFile=false Quiet=false
```

## Preconditions

- `NoLogFile = false` (default write file).
- Mock reaches complete successfully.

## Steps

1. Set `NoLogFile = false` explicitly.
2. Run full success path with default info logging.

## Context

- Requirement #1: optional `browser-trace.log` exists with similar milestone content.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.NoLogFile = false
	return nil
}
```
