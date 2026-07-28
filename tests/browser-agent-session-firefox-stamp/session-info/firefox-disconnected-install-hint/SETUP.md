# Scenario

**Feature**: disconnected session info recommends Firefox install for firefox sessions

```
SessionNew(Browser=firefox, NoOpen) -> session info
  -> Next steps mention install-firefox-extension and/or about:debugging
  -> path contains browser-agent-firefox
  -> not primary chrome://extensions / install-chrome-extension only
```

## Preconditions

- SessionInfoOp = firefox-disconnected-install-hint.

## Steps

1. Set SessionInfoOp = SessionInfoOpFirefoxDisconnectedInstallHint.
2. Set Browser = "firefox".

## Context

- Human CLI path exercises FormatSessionInfo against live snap.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionInfoOp = SessionInfoOpFirefoxDisconnectedInstallHint
	req.Browser = "firefox"
	req.SessionID = "sess-ff-stamp-info-1"
	return nil
}
```
