# Scenario

**Feature**: --no-color disables ANSI on install-firefox-extension stdout

```
HandleCLI install-firefox-extension --no-color (HOME=TestHome)
  -> stdout has no ESC sequences; path markers present
```

## Preconditions

- InstallCLIOp = color-no-color-flag.
- TestHome isolates extract.

## Steps

1. Set InstallCLIOp = InstallCLIOpColorNoColorFlag.

## Context

- Force-off must strip color even if TTY; on pipe both auto and force-off are plain.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.InstallCLIOp = InstallCLIOpColorNoColorFlag
	return nil
}
```
