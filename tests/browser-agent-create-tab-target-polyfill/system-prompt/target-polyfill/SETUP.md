# Scenario

**Feature**: FormatSystemPrompt describes Target.* polyfill + tab_id (C2)

```
FormatSystemPrompt(...)
  -> Target polyfilled / chrome.tabs lifecycle
  -> results use tab_id
  -> not only "Forbidden Target.*" / -32000 Not allowed as sole guidance
```

## Preconditions

- Mode already system-prompt from parent.

## Steps

1. Ensure ModeSystemPrompt.

## Context

- Requirement C2. Replaces blanket Forbidden Target.* wording.

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
