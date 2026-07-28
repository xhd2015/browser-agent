# Scenario

**Feature**: Browser=firefox create path stamps firefox identity

```
SessionNew(Browser=firefox) -> meta/snapshot firefox path + browser field
```

## Preconditions

- SessionNewBrowser = firefox.
- Browser = firefox.

## Steps

1. Set SessionNewBrowser = SessionNewBrowserFirefox.
2. Set Browser = "firefox".

## Context

- Child leaves pick meta vs v1 observation channel.

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
