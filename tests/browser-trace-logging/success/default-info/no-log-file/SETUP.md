# Scenario

**Feature**: default info milestones on stderr only — no session log file

```
# NoLogFile suppresses file mirror; stderr still has info
Lifecycle Logger -> stderr (info milestones)
Lifecycle Logger -/-> browser-trace.log
NoLogFile=true
```

## Preconditions

- `NoLogFile = true`.
- Default verbosity (not Quiet).

## Steps

1. Set `NoLogFile = true`.
2. Run success path; assert no log file artifact.

## Context

- Covers `--no-log-file` / `Config.NoLogFile` without changing stderr info.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.NoLogFile = true
	return nil
}
```
