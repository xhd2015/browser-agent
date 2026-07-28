# Scenario

**Feature**: CLI session new --browser firefox with inject OpenFirefoxFn

```
HandleCLI ["session","new","--browser","firefox", flags...]
  + WithSessionNewHooks(OpenFirefoxFn)
  -> exit 0; OpenFirefoxFn once; stdout path + about:debugging
```

## Preconditions

- CLIOp = session-new-browser-firefox.
- HOME env = TestHome for ensure isolation on CLI path.
- Ephemeral daemon via RunDaemon; --host/--server-port + --no-wait.

## Steps

1. Set CLIOp = CLIOpSessionNewBrowserFirefox.
2. Set Browser = "firefox".

## Context

- OpenChromeFn inject returns error if called (must not run).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpSessionNewBrowserFirefox
	req.Browser = "firefox"
	return nil
}
```
