# Scenario

**Feature**: --no-open-chrome with Browser=firefox skips open but still prints path

```
SessionNew(Browser=firefox, NoOpenChrome=true)
  -> OpenFirefoxFn never called
  -> stdout still has browser-agent-firefox path + about:debugging
```

## Preconditions

- SessionNewFirefoxOp = no-open-still-prints-path.
- NoOpenChrome forced true.

## Steps

1. Set SessionNewFirefoxOp = SessionNewFirefoxOpNoOpenStillPrintsPath.
2. Set NoOpenChrome = true.

## Context

- Manual-load path: operator still needs the printed extension path.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionNewFirefoxOp = SessionNewFirefoxOpNoOpenStillPrintsPath
	req.NoOpenChrome = true
	return nil
}
```
