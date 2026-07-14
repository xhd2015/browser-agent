# Scenario

**Feature**: FormatSystemPrompt mentions BROWSER_AGENT_SESSION_ID env source (B3)

```
FormatSystemPrompt("any")
  -> contains "BROWSER_AGENT_SESSION_ID"
```

## Preconditions

- Any session id string (must not appear in body — not asserted here).

## Steps

1. Set PromptSessionID.

## Context

- Requirement B3.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PromptSessionID = "sess-env-mention"
	return nil
}
```
