# Scenario

**Feature**: SessionNew with Browser=firefox

```
SessionNew(Browser=firefox, Home=TestHome, OpenFirefoxFn, NoWait)
  -> EnsureCanonicalFirefoxExtensionWithHome
  -> OpenFirefoxFn(sessionURL) unless NoOpenChrome
  -> stdout: firefox path + about:debugging
```

## Preconditions

- SessionNewBrowser = firefox.
- TestHome isolation for ensure path.

## Steps

1. Set SessionNewBrowser = SessionNewBrowserFirefox.
2. Set Browser = "firefox".

## Context

- OpenChromeFn is still injected and must not be called.
- No real Firefox process.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionNewBrowser = SessionNewBrowserFirefox
	req.Browser = "firefox"
	return nil
}
```
