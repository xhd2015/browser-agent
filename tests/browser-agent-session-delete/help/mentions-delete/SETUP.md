# Scenario

**Feature**: help documents session delete sub-command

```
HandleCLI session --help -> fullHelp lists session delete
```

## Preconditions

- Read-only help dispatch.

## Steps

1. `HelpArgs = ["session", "--help"]`.

## Context

- Current help lacks delete — leaf is **RED** until implementer updates `fullHelp`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HelpArgs = []string{"session", "--help"}
	return nil
}
```