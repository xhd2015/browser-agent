# Scenario

**Feature**: Extension popup + background progressive status contracts (P4)

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent/public/popup.html, popup.js, background.js
Test Client -> assert instant shell, health timeout, phase hooks,
               SW status API, last-known paint
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; filesystem reads only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Leaf sets `ExtSourceTarget`.

## Context

- Classic TDD: health timeout, progressive phases, SW status API, and last-known
  leaves are **RED** under current health-only popup until implementer lands P4.
- `shell-version-instant` may already **GREEN** (sync bundle-sum version paint).
- MECE with attach-gate / eager-arm / self-heal — do not re-assert those here.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeExtSource
	return nil
}
```
