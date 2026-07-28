# Scenario

**Feature**: --color forces ANSI on install-firefox-extension stdout (pipe)

```
HandleCLI install-firefox-extension --color (HOME=TestHome)
  -> stdout has ESC[32m and/or ESC[33m; path markers present
```

## Preconditions

- InstallCLIOp = color-force-on.
- Explicit env without NO_COLOR; TestHome isolates extract.

## Steps

1. Set InstallCLIOp = InstallCLIOpColorForceOn.

## Context

- Pipe is non-TTY; `--color` must opt in to coloring (green success / yellow warning).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.InstallCLIOp = InstallCLIOpColorForceOn
	return nil
}
```
