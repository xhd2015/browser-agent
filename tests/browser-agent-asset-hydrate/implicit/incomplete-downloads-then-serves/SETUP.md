# Scenario

**Feature**: incomplete embed triggers ensure then serves from cache

```
embedFS = empty
httptest serves session-page tar.gz
  -> ResolveSessionPage(empty, v0.2.0, cfg)
  -> html non-empty, source == "cache"
  -> GETCount >= 1
  -> CacheComplete(session-page) true
```

## Preconditions

- Empty embed fixture.
- XDG temp cold.
- httptest serves complete session-page archive.

## Steps

1. Set `ImplicitOp = ImplicitOpIncompleteDownloadsThenServes`.
2. Embed empty; XDG temp; serve fixture session-page-complete.

## Context

- go-install incomplete embed path for session page.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ImplicitOp = ImplicitOpIncompleteDownloadsThenServes
	req.ImplicitEmbedFixture = FixtureEmpty
	req.XDGCacheHome = t.TempDir()
	req.ImplicitVersion = CacheVersion
	req.ImplicitServeFixture = FixtureSessionPageComplete
	return nil
}
```
