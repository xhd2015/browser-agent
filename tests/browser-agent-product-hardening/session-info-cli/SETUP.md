# Scenario

**Feature**: session info CLI distinguishes control snapshot vs browser inventory

```
HandleCLI(["session","info", …])
  -> control GET /v1/session always
  -> browser info job only when extension.connected
  -> missing session id → flag+env error
```

## Preconditions

- Mode = session-info-cli for this subtree.
- Live leaves use in-process serve (NoOpenChrome, NoAgentRun).
- CLIEnv empty map by default (ignore ambient process env).

## Steps

1. Set Mode to ModeSessionInfoCLI.
2. Default CLIEnv to empty map when nil.
3. Child leaves set SessionInfoKind.

## Context

- Requirement B1–B3. Prefer package HandleCLI over building cmd binary.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionInfoCLI
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	return nil
}
```
