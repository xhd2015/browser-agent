# Scenario

**Feature**: typo flag on serve subcommand

```
HandleCLI serve --foo -> unrecognized flag --foo -> exit 1
```

## Preconditions

- Default args `serve --foo --base-dir <temp>` unless overridden.

## Steps

1. Leaf uses default unknown-flag args from Run.

## Context

- Covers less-flags fatal parse path for serve.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```