# Scenario

**Feature**: Verbose mode adds detail (hello, start recording, stop, complete)

```
# Verbose extras on top of (or instead of purely) info
browser-trace Verbose=true Quiet=false
Lifecycle Logger -> stderr mentions hello and/or version
```

## Preconditions

- `Verbose = true`, `Quiet = false`.
- Mock posts hello with a version string.

## Steps

1. Enable Verbose.
2. Keep Quiet off.

## Context

- Requirement scenario #4.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Verbose = true
	req.Quiet = false
	return nil
}
```
