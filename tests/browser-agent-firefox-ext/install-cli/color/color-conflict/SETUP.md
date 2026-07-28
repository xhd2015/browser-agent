# Scenario

**Feature**: --color and --no-color cannot be used together on install-firefox-extension

```
HandleCLI install-firefox-extension --color --no-color
  -> fatal error exit 1; cannot be specified together
```

## Preconditions

- InstallCLIOp = color-conflict.
- Both color flags on the same invocation.
- TestHome set (extract not required if flags fail before install; still isolated).

## Steps

1. Set InstallCLIOp = InstallCLIOpColorConflict.

## Context

- Same mutual exclusion message as serve / session list color flags.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.InstallCLIOp = InstallCLIOpColorConflict
	return nil
}
```
