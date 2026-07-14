# Scenario

**Feature**: CLI help documents session list sub-command

```
HandleCLI session --help -> fullHelp lists session list
```

## Preconditions

- Mode `help`; read-only CLI dispatch.

## Steps

1. Set `Mode = ModeHelp`.
2. Default `HelpArgs = ["session", "--help"]` when empty.

## Context

- Help must mention `list` under session subcommands.

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
}```
