# Scenario

**Feature**: neither flag nor env → hard error naming both (A3 / C4)

```
ResolveSessionID(flag unset, env unset) -> error
  message mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- FlagSet=false, EnvSet=false.

## Steps

1. Clear flag and env presence.

## Context

- Error text is user-facing; both sources must be mentioned.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FlagSet = false
	req.FlagValue = ""
	req.EnvSet = false
	req.EnvValue = ""
	return nil
}
```
