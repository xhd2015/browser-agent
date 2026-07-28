# Scenario

**Feature**: default chrome SessionNew must not force firefox stamp

```
SessionNew(Browser="") -> chrome extension path (not browser-agent-firefox)
```

## Preconditions

- SessionNewBrowser = chrome (default empty Browser).

## Steps

1. Set SessionNewBrowser = SessionNewBrowserChrome.
2. Set Browser = "" (chrome default).

## Context

- Regression guard so firefox stamp does not break chrome create.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionNewBrowser = SessionNewBrowserChrome
	req.Browser = ""
	return nil
}
```
