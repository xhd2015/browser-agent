# Scenario

**Feature**: forced browser prop wins over EXT / UA / snap / boot (P1)

```
resolveInstallBrowser
  # forced === "firefox" | "chrome"
  -> return forced first (before EXT / UA / boot)
```

## Preconditions

- Mode already `ModeReactSrc` from parent.

## Steps

1. Set `ReactProbe = ReactProbeResolveForcedProp`.

## Context

- Early return on forced prop is highest priority.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeResolveForcedProp
	return nil
}
```
