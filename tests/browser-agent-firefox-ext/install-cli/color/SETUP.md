# Scenario

**Feature**: cli-color on install-firefox-extension stdout

```
HandleCLI install-firefox-extension --color|--no-color
  -> ANSI green/yellow on force-on; plain on --no-color; conflict on both
```

## Preconditions

- Mode already ModeInstallCLI.
- TestHome isolates extract path (setenv HOME for HandleCLI).
- Non-TTY pipe / buffer: `--color` must force ANSI on.

## Steps

1. Leaf sets InstallCLIOp to a color-* probe.

## Context

- Color is applied to install help **stdout** (not serve stderr).
- Conflict returns nonzero with "cannot be specified together".

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// Mode already set by install-cli parent; color leaves only narrow InstallCLIOp.
	if req.Env == nil {
		req.Env = map[string]string{}
	}
	return nil
}
```
