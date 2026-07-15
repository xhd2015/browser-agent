# Scenario

**Feature**: `assets --help` documents ensure and status

```
HandleCLI(["assets", "--help"], …)
  -> nil error
  -> stdout mentions ensure and status
  -> trailing \n
```

## Preconditions

- Args: `assets --help` (assets -h / assets help also acceptable product-wide;
  this leaf pins `--help`).

## Steps

1. Set `CLIOp = CLIOpAssetsHelp`.
2. Set CLIArgs to `assets --help`.

## Context

- Help must not start capture session or network.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpAssetsHelp
	req.CLIArgs = []string{"assets", "--help"}
	req.CLIEnv = map[string]string{}
	return nil
}
```
