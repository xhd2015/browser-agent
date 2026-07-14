# Scenario

**Feature**: serve --session-id prints deprecation warning on stderr

```
HandleCLI(serve --session-id sess-cli-warn ...) -> stderr deprecation
```

## Preconditions

- `--no-open-chrome`, `--no-agent-run` for fast start.

## Steps

1. Set `CLISessionID = "sess-cli-warn"`.

## Context

- Warning must appear before blocking serve; harness polls stderr ≤3s.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLISessionID = "sess-cli-warn"
	return nil
}
```