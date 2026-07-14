# Scenario

**Feature**: --help lists serve, info, eval (A2)

```
HandleCLI(["--help"])
  -> help text lists serve, info, eval
  -> nil error (no os.Exit)
  -> trailing \n
```

## Preconditions

- CLIArgs = ["--help"] (also valid product may accept -h; this leaf uses --help).

## Steps

1. Set CLIArgs to `--help`.
2. Empty CLIEnv.

## Context

- Help must not hang in serve.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIArgs = []string{"--help"}
	req.CLIEnv = map[string]string{}
	return nil
}
```
