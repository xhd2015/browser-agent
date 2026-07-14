# Scenario

**Feature**: FormatSystemPrompt nested recipes without concrete control id

```
Test Client -> FormatSystemPrompt(sessionID) -> playbook text
# browser-agent session info|eval|run|logs|screenshot|cdp
# no concrete session id argument; mentions BROWSER_AGENT_SESSION_ID
```

## Preconditions

- Mode is system-prompt.
- PromptSessionID set by leaves (unique for no-control-id leaf).

## Steps

1. Set Mode = ModeSystemPrompt.
2. Leave PromptSessionID to leaf.

## Context

- API may still accept sessionID for signature stability but must not embed it.

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
