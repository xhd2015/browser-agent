# Scenario

**Feature**: EnsureCanonicalExtension writes manifest under browser-agent segment

```
EnsureCanonicalExtension() -> .../extensions/browser-agent/{ver}/manifest.json
```

## Preconditions

- Fresh `TestHome`.

## Steps

1. Set `CanonicalPathOp = extracts-under-browser-agent`.

## Context

- Path must be absolute.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.CanonicalPathOp = CanonicalPathOpExtractsUnderBrowserAgent
	return nil
}
```