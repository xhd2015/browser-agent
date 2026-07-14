# Scenario

**Feature**: verbose success surfaces hello/version on stderr

```
# Mock hello carries version test-mock-1.0.0
Mock Extension -> POST /v1/hello {version}
Lifecycle Logger Verbose -> stderr: hello and/or version
browser-trace stdout -> "{sessionDir}\n"
```

## Preconditions

- Verbose from parent; mock uses record-and-complete (posts hello first).

## Steps

1. Keep Verbose true.
2. Run success path.

## Context

- Requirement #4: stderr mentions hello and/or version when Verbose.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Verbose = true
	req.Quiet = false
	req.NoLogFile = false
	return nil
}
```
