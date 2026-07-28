# Scenario

**Feature**: Firefox ensure path uses browser-agent-firefox segment (not managed-chrome)

```
EnsureCanonicalFirefoxExtensionWithHome(TestHome)
  -> .../extensions/browser-agent-firefox/{ver}/manifest.json
```

## Preconditions

- EnsureCanonicalOp = path-under-browser-agent-firefox.
- Fresh TestHome.

## Steps

1. Set EnsureCanonicalOp = EnsureCanonicalOpPathUnderFirefox.

## Context

- Path must be absolute; must not nest under managed-chrome.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.EnsureCanonicalOp = EnsureCanonicalOpPathUnderFirefox
	return nil
}
```
