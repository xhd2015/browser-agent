# Scenario

**Feature**: browser-agent `assets` CLI via HandleCLI (P5)

```
HandleCLI(["assets", ...], env, stdout, stderr)
  -> help | status | ensure
```

## Preconditions

- Mode is cli.
- Leaf sets CLIOp, CLIArgs, CLIEnv / XDG / archive server flags.

## Steps

1. Set `Mode = ModeCLI`.

## Context

- Classic TDD for P5 — RED until assets subcommand exists on HandleCLI.
- Prefer package HandleCLI over building the binary.

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
