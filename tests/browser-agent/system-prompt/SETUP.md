# Scenario

**Feature**: SYSTEM.md formatter and product default port

```
# Pure package — no Chrome, no agent-run
Test Client -> FormatSystemPrompt(sessionID) -> playbook text
# nested session recipes; no concrete control id
Test Client -> DefaultAddr / DefaultControlPort -> 43761
```

## Preconditions

- Mode is `system-prompt`.
- No server required.

## Steps

1. Set `Mode = ModeSystemPrompt`.
2. Children set SystemOp.

## Context

- Requirement F1–F2. On-disk write during serve is implementer detail; pure
  formatter is the GREEN contract for this tree.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSystemPrompt
	return nil
}
```
