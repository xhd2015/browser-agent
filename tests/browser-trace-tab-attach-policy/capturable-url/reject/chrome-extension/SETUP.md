# Scenario

**Feature**: `chrome-extension://` pages are not capturable (requirement #3)

```
IsCapturableTabURL("chrome-extension://abcdefghijklmnopqrstuvwxyz123456/popup.html") -> false
ShouldAttemptAttach(open gates, chrome-extension URL) -> false
```

## Preconditions

- Extension pages (including this capture extension) must not receive attach.
- Gates open.

## Steps

1. Set `URL` to a representative `chrome-extension://…` page URL.

## Context

- Matches existing `attachAllTabsInWindow` skip: `url.startsWith('chrome-extension://')`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.URL = "chrome-extension://abcdefghijklmnopqrstuvwxyz123456/popup.html"
	return nil
}
```
