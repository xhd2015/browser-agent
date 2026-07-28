# Scenario

**Feature**: EnsureCanonicalFirefoxExtensionWithHome writes Firefox package under home

```
TestHome -> EnsureCanonicalFirefoxExtensionWithHome
  -> {TestHome}/.browser-agent/extensions/browser-agent-firefox/{version}/
```

## Preconditions

- Mode = ensure-canonical.
- TestHome isolated temp from root Setup.
- Not under managed-chrome.

## Steps

1. Set Mode = ModeEnsureCanonical.
2. Leaf sets EnsureCanonicalOp.

## Context

- Version from package/fixture manifest; idempotent per version.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeEnsureCanonical
	return nil
}
```
