# Scenario

**Feature**: session new help documents --browser

```
HandleCLI session new --help -> stdout mentions --browser (chrome|firefox)
```

## Preconditions

- CLIOp = help-mentions-browser.
- No daemon required.

## Steps

1. Set CLIOp = CLIOpHelpMentionsBrowser.

## Context

- fullHelp / session new --help must document the flag for operators.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpHelpMentionsBrowser
	return nil
}
```
