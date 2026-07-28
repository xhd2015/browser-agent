# Scenario

**Feature**: chrome create keeps Chrome extension_install_path (not firefox)

```
SessionNew(default chrome) -> meta.extension_install_path is browser-agent (Chrome), not browser-agent-firefox
```

## Preconditions

- SessionNewChromeOp = meta-not-firefox-path.

## Steps

1. Set SessionNewChromeOp = SessionNewChromeOpMetaNotFirefoxPath.
2. Use distinct SessionID to avoid collision with firefox leaves in shared docs only
   (per-leaf BaseDir already isolates).

## Context

- Chrome path may live under managed-chrome; must not be firefox segment.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionNewChromeOp = SessionNewChromeOpMetaNotFirefoxPath
	req.SessionID = "sess-ff-stamp-chrome-1"
	return nil
}
```
