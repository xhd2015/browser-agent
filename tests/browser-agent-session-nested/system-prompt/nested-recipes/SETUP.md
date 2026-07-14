# Scenario

**Feature**: FormatSystemPrompt includes nested session CLI recipes (B1)

```
FormatSystemPrompt("sess-nested-recipes")
  -> contains "browser-agent session info"
  -> contains session eval/run/logs/screenshot/cdp recipes
```

## Preconditions

- PromptSessionID fixed for stability (not asserted present in body).

## Steps

1. Set PromptSessionID.

## Context

- Requirement B1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PromptSessionID = "sess-nested-recipes"
	return nil
}
```
