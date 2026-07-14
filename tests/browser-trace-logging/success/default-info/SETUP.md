# Scenario

**Feature**: default log mode — info milestones on stderr (not Quiet, not Verbose)

```
# Default verbosity: info lifecycle milestones
Lifecycle Logger (default) -> stderr: listen, session, ready, recording, …
browser-trace Verbose=false Quiet=false
```

## Preconditions

- `Verbose = false`, `Quiet = false`.
- Descendants split only on log-file policy (`NoLogFile`).

## Steps

1. Force default verbosity flags off.
2. Leave `NoLogFile` for leaf override.

## Context

- Requirement scenario #1 base (stderr milestones + clean stdout).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Verbose = false
	req.Quiet = false
	return nil
}
```
