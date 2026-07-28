# Scenario

**Feature**: Go inject/fallback still has Chrome install path (G2)

```
browseragent/server.go injectSessionBoot / writeFallbackSessionHTML
  chrome://extensions remains for Chrome / default browser path
```

## Preconditions

- Mode already go-src from parent.
- Out of scope: removing Chrome default install path.

## Steps

1. Set `GoSrcProbe = GoSrcChromeInstallPreserved`.

## Context

- Expect **GREEN** already (current inject is Chrome-primary).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.GoSrcProbe = GoSrcChromeInstallPreserved
	return nil
}
```
