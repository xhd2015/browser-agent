# Scenario

**Feature**: complete embed FS serves without download

```
embedFS = testdata/session-page-complete
  -> ResolveSessionPage(embedFS, v0.2.0, cfg)
  -> html non-empty, source == "embed"
  -> GETCount == 0 (httptest never hit)
```

## Preconditions

- Complete session-page fixture as embedFS.
- XDG temp (unused on happy path).
- Server available only to prove zero GETs.

## Steps

1. Set `ImplicitOp = ImplicitOpCompleteEmbedNoDownload`.
2. Set embed fixture complete; XDG temp; start counter server in Run.

## Context

- Runtime path when fat release embed is complete.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ImplicitOp = ImplicitOpCompleteEmbedNoDownload
	req.ImplicitEmbedFixture = FixtureSessionPageComplete
	req.XDGCacheHome = t.TempDir()
	req.ImplicitVersion = CacheVersion
	req.ImplicitStartServer = true
	return nil
}
```
