# Scenario

**Feature**: CLI help strings in `browseragent/cli.go`

```
HandleCLI --help | serve --help | session --help
  -> fullHelp / briefUsage contract markers
```

## Preconditions

- Mode `ModeCLIHelp`.
- Leaf sets `CLIHelpProbe`.

## Steps

1. Set `Mode = ModeCLIHelp`.

## Context

- Read-only; no subprocess shell-out to binary.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLIHelp
	return nil
}
```