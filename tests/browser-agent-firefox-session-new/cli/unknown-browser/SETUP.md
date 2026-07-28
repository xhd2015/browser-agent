# Scenario

**Feature**: unknown --browser value errors with nonzero exit

```
HandleCLI session new --browser safari … -> error containing "unknown browser"; nonzero
```

## Preconditions

- CLIOp = unknown-browser.
- UnknownBrowser = "safari".
- --no-open-chrome --no-wait so open path is not required if validation is late.

## Steps

1. Set CLIOp = CLIOpUnknownBrowser.
2. Set UnknownBrowser = "safari".

## Context

- Prefer fail-fast validation before EnsureDaemon; test does not require healthy daemon.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpUnknownBrowser
	req.UnknownBrowser = "safari"
	return nil
}
```
