# Scenario

**Feature**: FormatSystemPrompt nested CLI recipes without control id (F1)

```
FormatSystemPrompt("sess-system-prompt") contains:
  browser-agent session info/eval/run/logs/screenshot
  BROWSER_AGENT_SESSION_ID
  (does not embed concrete session id)
```

## Preconditions

- SystemOp format.
- PromptSessionID fixed for assert (absence check).

## Steps

1. Set SystemOp to format.
2. Set PromptSessionID `sess-system-prompt`.

## Context

- Nested recipes match complete session CLI refactor.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SystemOp = SystemOpFormat
	req.PromptSessionID = "sess-system-prompt"
	return nil
}
```
