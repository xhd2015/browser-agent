# Scenario

**Feature**: popup shell shows package version without waiting on network

```
popup.html opens
  -> #pkg-version (or equivalent) present in shell
  -> version from bundle-sum.js / BROWSER_AGENT_BUNDLE_VERSION (sync)
  -> not only set after fetch(/v1/health).then(...)
```

## Preconditions

- ExtSourceTarget = shell-version-instant.

## Steps

1. Set `ExtSourceTarget = ExtSrcShellVersionInstant`.

## Context

- Current product already paints version via sync `bundle-sum.js` + `#pkg-version`
  — this leaf is expected **GREEN** as a regression guard for instant shell.
- Does not require progressive phase rows (covered by progressive-status-phases).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcShellVersionInstant
	return nil
}
```
