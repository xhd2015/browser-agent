# Scenario

**Feature**: CLI `session new` creates session end-to-end

```
HandleCLI session new --base-dir --addr --session-id -> exit 0 + session on server
```

## Preconditions

- Explicit session id `sess-cli-8`.
- `CLIDispatchOp` session-new-subcommand.

## Steps

1. Set `CLIDispatchOp = CLIDispatchOpSessionNewSubcommand`.
2. Set `SessionID = "sess-cli-8"`.

## Context

- `HandleCLI` must return nil (exit 0).
- Stdout should include session id and nested session recipes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIDispatchOp = CLIDispatchOpSessionNewSubcommand
	req.SessionID = "sess-cli-8"
	return nil
}```
