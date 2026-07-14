# Scenario

**Feature**: Create writes SYSTEM.md from FormatSystemPrompt

```
Create(id) -> SYSTEM.md content == FormatSystemPrompt(id)
```

## Preconditions

- CreateCase ok; SessionID `playbook-test`.

## Steps

1. Set SessionID to `playbook-test` for distinct artifact naming.

## Context

- FormatSystemPrompt keeps session id out of body; content is still deterministic.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "playbook-test"
	return nil
}
```