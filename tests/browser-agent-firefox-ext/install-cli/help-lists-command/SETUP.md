# Scenario

**Feature**: top-level --help lists install-firefox-extension

```
HandleCLI(--help) -> stdout mentions install-firefox-extension
```

## Preconditions

- InstallCLIOp = help-lists-command.

## Steps

1. Set InstallCLIOp = InstallCLIOpHelpListsCommand.

## Context

- install-chrome-extension must still be listed (do not remove Chrome command).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.InstallCLIOp = InstallCLIOpHelpListsCommand
	return nil
}
```
