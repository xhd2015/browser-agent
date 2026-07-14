# Scenario

**Feature**: FormatSystemPrompt body omits concrete control session id (B2)

```
FormatSystemPrompt("ctrl-sysmd-unique-qq7")
  -> body does NOT contain "ctrl-sysmd-unique-qq7"
```

## Preconditions

- Unique control id that will not appear as incidental product text.

## Steps

1. Set PromptSessionID to unique token.

## Context

- Requirement B2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PromptSessionID = "ctrl-sysmd-unique-qq7"
	return nil
}
```
