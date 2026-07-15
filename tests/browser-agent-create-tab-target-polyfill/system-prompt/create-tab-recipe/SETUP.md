# Scenario

**Feature**: FormatSystemPrompt includes session create-tab recipe (C1)

```
FormatSystemPrompt("sess-create-tab-prompt")
  -> contains browser-agent session create-tab
  -> does not embed concrete session id
```

## Preconditions

- Mode already system-prompt from parent.

## Steps

1. Keep default PromptSessionID (or set create-tab marker sid).

## Context

- Requirement C1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSystemPrompt
	req.PromptSessionID = "sess-create-tab-prompt"
	return nil
}
```
