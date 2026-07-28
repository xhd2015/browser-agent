# Scenario

**Feature**: inject/fallback Firefox install markers (G1)

```
browseragent/server.go injectSessionBoot / writeFallbackSessionHTML
  (and/or install-panel helpers they call)
  when browser is firefox:
    about:debugging
    temporary add-on / Load Temporary Add-on
    browser-agent-firefox
```

## Preconditions

- Mode already go-src from parent.

## Steps

1. Set `GoSrcProbe = GoSrcInjectFirefoxInstall`.

## Context

- Markers may live in a shared helper used by inject + fallback.
- CLI-only about:debugging in extension_firefox.go is **not** enough alone
  unless session page inject/fallback also references Firefox install UX.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.GoSrcProbe = GoSrcInjectFirefoxInstall
	return nil
}
```
