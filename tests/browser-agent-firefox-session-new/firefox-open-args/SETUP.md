# Scenario

**Feature**: pure BuildFirefoxOpenArgs (URL-only; no managed Firefox flags)

```
Test Client -> BuildFirefoxOpenArgs(sessionURL) -> argv without binary
  # never --load-extension / --user-data-dir
```

## Preconditions

- Mode = firefox-open-args.
- Package exports `BuildFirefoxOpenArgs` (RED until implementer).

## Steps

1. Set Mode = ModeFirefoxOpenArgs.

## Context

- Pure function; no daemon; no browser process.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeFirefoxOpenArgs
	return nil
}
```
