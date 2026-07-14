# Scenario

**Feature**: FormatSystemPrompt contains all six nested session CLI recipes (C1)

```
FormatSystemPrompt("sess-cdp-prompt")
  -> non-empty
  -> does not embed concrete session id
  -> contains browser-agent session info|eval|run|logs|screenshot|cdp
  -> mentions BROWSER_AGENT_SESSION_ID
```

## Preconditions

- PromptSessionID = sess-cdp-prompt (absence asserted).

## Steps

1. Set PromptSessionID.

## Context

- Requirement C1 — nested complete refactor.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PromptSessionID = "sess-cdp-prompt"
	return nil
}
```
