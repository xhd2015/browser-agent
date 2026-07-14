# Scenario

**Feature**: CLI help documents session delete sub-command

```
HandleCLI session --help -> fullHelp lists session delete
```

## Preconditions

- Mode `help`; read-only CLI dispatch.

## Steps

1. Set `Mode = ModeHelp`.
2. Default `HelpArgs = ["session", "--help"]` when empty.

## Context

- Help must mention `delete` under session subcommands.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeHelp
	if len(req.HelpArgs) == 0 {
		req.HelpArgs = []string{"session", "--help"}
	}
	return nil
}
```