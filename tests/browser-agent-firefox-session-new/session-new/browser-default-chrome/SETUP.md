# Scenario

**Feature**: SessionNew default/chrome path unchanged when Browser empty

```
SessionNew(Browser="", OpenChromeFn, OpenFirefoxFn, NoWait)
  -> OpenChromeFn once; OpenFirefoxFn never
  -> stdout is chrome-oriented (not firefox-primary about:debugging install)
```

## Preconditions

- SessionNewBrowser = default-chrome.
- Browser field empty (chrome default).

## Steps

1. Set SessionNewBrowser = SessionNewBrowserDefaultChrome.
2. Clear Browser string.

## Context

- Smoke: chrome remains default when --browser omitted.
- Does not re-test full chrome open-chrome tree; only no-regression markers.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionNewBrowser = SessionNewBrowserDefaultChrome
	req.Browser = ""
	req.NoOpenChrome = false
	return nil
}
```
