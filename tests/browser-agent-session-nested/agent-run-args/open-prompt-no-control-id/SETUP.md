# Scenario

**Feature**: open prompt for SYSTEM.md path omits bare control id (A4)

```
BuildAgentRunArgs("ctrl-open-unique-zz9", "/tmp/browser-agent-playbook/SYSTEM.md", "")
  -> open prompt after "--" uses absolute SYSTEM.md path
  -> does NOT contain bare control id "ctrl-open-unique-zz9"
```

## Preconditions

- Unique control id unlikely to appear as incidental substring.
- Prompt path ends with SYSTEM.md (triggers playbook open prompt).

## Steps

1. Set AgentArgsControlID to unique token.
2. Set SYSTEM.md path; empty workspace.

## Context

- Requirement A4: open prompt / SYSTEM path must not require bare control id.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AgentArgsControlID = "ctrl-open-unique-zz9"
	// Absolute playbook path must NOT include the control id (id is only via --env).
	req.AgentArgsPromptPath = "/tmp/browser-agent-playbook/SYSTEM.md"
	req.AgentArgsWorkspace = ""
	return nil
}
```
