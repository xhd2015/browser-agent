# Scenario

**Feature**: HandleCLI nested session dispatch (help / resolve / flat unknown / eval)

```
Operator -> HandleCLI(["--help"] | ["session", …] | ["info"])
  -> help lists session
  -> session info without sid → flag+env error
  -> flat info → unknown / non-success
  -> session eval + fake WS → job type eval
```

## Preconditions

- Mode is cli-nested.
- CLIEnv empty map so ambient process env is ignored.
- No real Chrome / agent-run.

## Steps

1. Set Mode = ModeCLINested.
2. Default empty CLIEnv and MaxDispatchWait.
3. Leave CLIKind / CLIArgs to leaf.

## Context

- Complete refactor: no flat side-command aliases.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLINested
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 3 * time.Second
	}
	return nil
}
```
