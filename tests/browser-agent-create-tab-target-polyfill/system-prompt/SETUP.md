# Scenario

**Feature**: FormatSystemPrompt documents create-tab + Target polyfill

```
Test Client -> FormatSystemPrompt(sessionID) -> playbook text
  must mention browser-agent session create-tab
  must describe Target.* as polyfilled (tab_id), not Forbidden-only
```

## Preconditions

- Mode is system-prompt.
- Pure call; no server.

## Steps

1. Set `Mode = ModeSystemPrompt`.
2. Default PromptSessionID when leaf omits it.

## Context

- Requirements C1–C2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSystemPrompt
	if req.PromptSessionID == "" {
		req.PromptSessionID = "sess-create-tab-prompt"
	}
	return nil
}
```
