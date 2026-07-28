# Scenario

**Feature**: SessionNew package API with Browser selection

```
SessionNew(Browser, OpenFirefoxFn|OpenChromeFn, Home, NoWait)
  -> create session + optional open + pretty stdout
```

## Preconditions

- Mode = session-new.
- Ephemeral daemon + isolated BaseDir / TestHome from root Setup.
- NoWait always true (no real extension connection).

## Steps

1. Set Mode = ModeSessionNew.

## Context

- Package API only (no HandleCLI here).
- Open hooks are local closures (parallel-safe; no bare globals).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionNew
	return nil
}
```
