# Scenario

**Feature**: FormatSystemPrompt documents full CLI recipe surface including cdp

```
Test Client -> FormatSystemPrompt(sessionID) -> playbook text
  must mention browser-agent info|eval|run|logs|screenshot|cdp
```

## Preconditions

- Mode is system-prompt.
- Pure call; no server.

## Steps

1. Set `Mode = ModeSystemPrompt`.
2. Default PromptSessionID when leaf omits it.

## Context

- Requirement C1. Distinct from sealed system-prompt leaf that may omit cdp assert.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSystemPrompt
	if req.PromptSessionID == "" {
		req.PromptSessionID = "sess-cdp-prompt"
	}
	return nil
}
```
