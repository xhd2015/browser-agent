# Scenario

**Feature**: less-flags unknown serve flag errors

```
HandleCLI serve --foo -> unrecognized flag -> exit 1
```

## Preconditions

- Mode `ModeUnknownFlag`.
- less-flags parse on `cliServe`.

## Steps

1. Set `Mode = ModeUnknownFlag`.

## Context

- Error should suggest `browser-agent serve --help`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeUnknownFlag
	return nil
}
```