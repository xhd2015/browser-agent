# Scenario

**Feature**: bare HandleCLI([]) prints brief usage (A1)

```
HandleCLI([])
  -> print brief usage (stdout or stderr)
  -> mentions serve
  -> non-nil error
  -> trailing \n on printed body
```

## Preconditions

- CLIArgs empty/nil.
- Must not start long-running serve.

## Steps

1. Set CLIArgs to empty slice.
2. Ensure CLIEnv has no session id.

## Context

- Exact wording flexible; must stay short and name `serve`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIArgs = []string{}
	req.CLIEnv = map[string]string{}
	return nil
}
```
