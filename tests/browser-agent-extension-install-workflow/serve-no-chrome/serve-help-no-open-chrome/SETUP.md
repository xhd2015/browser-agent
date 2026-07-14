# Scenario

**Feature**: serve --help omits --no-open-chrome flag

```
HandleCLI(serve --help) -> help text without --no-open-chrome
```

## Preconditions

- Flag removed or hidden from operator help.

## Steps

1. Set `ServeNoChromeOp = serve-help-no-open-chrome`.

## Context

- Serve no longer launches chrome; flag is obsolete.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ServeNoChromeOp = ServeNoChromeOpServeHelpNoOpenChrome
	return nil
}
```