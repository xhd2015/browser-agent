# Scenario

**Feature**: install-firefox-extension stdout shows path and about:debugging steps

```
InstallFirefoxExtensionWithHome(stdout, TestHome)
  -> extensions/browser-agent-firefox/ + about:debugging + Load Temporary
```

## Preconditions

- InstallCLIOp = stdout-path-and-about-debugging.
- TestHome set by root Setup.

## Steps

1. Set InstallCLIOp = InstallCLIOpStdoutPathAndAboutDebugging.

## Context

- Temporary add-on unload-on-restart note is optional one line.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.InstallCLIOp = InstallCLIOpStdoutPathAndAboutDebugging
	return nil
}
```
