# Scenario

**Feature**: complete FS resolve returns session-page index HTML from embed

```
testdata/session-page-complete
  -> ResolveSessionPageIndexFS(fs)
  -> html contains data-browser-agent-root / #root
  -> source == "embed"
  -> err == nil
```

## Preconditions

- Fixture is complete (index + assets JS).
- ExpectResolveOK true.

## Steps

1. Set `ResolveFixtureName = FixtureSessionPageComplete`.
2. Set `ExpectResolveOK = true`.

## Context

- Mirrors product path: when embed complete, index is served from embed.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ResolveFixtureName = FixtureSessionPageComplete
	req.ExpectResolveOK = true
	return nil
}
```
