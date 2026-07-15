# Scenario

**Feature**: browser-trace `assets` CLI via HandleCLI (P6)

```
browsertrace.HandleCLI(["assets", ...], env, stdout, stderr)
  -> help | ensure  (extension only for product browser-trace)
```

## Preconditions

- Mode is cli.
- Leaf sets CLIOp, CLIArgs, CLIEnv / XDG / archive server flags.

## Steps

1. Set `Mode = ModeCLI`.

## Context

- Classic TDD — RED until `assets` subcommand exists on `browsertrace.HandleCLI`.
- Prefer package HandleCLI over building the binary.
- Ensure covers **extension only** (no session-page for browser-trace).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLI
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	return nil
}
```
