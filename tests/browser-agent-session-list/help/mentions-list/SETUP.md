# Scenario

**Feature**: help documents session list sub-command

```
HandleCLI session --help -> fullHelp lists session list
```

## Preconditions

- Read-only help dispatch.

## Steps

1. `HelpArgs = ["session", "--help"]`.

## Context

- Current help lacks list — leaf is **RED** until implementer updates `fullHelp`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HelpArgs = []string{"session", "--help"}
	return nil
}```
