# Scenario

**Feature**: incomplete extension embed ensures via download to complete path

```
embedFS = empty (extension)
httptest serves extension tar.gz (manifest + background)
  -> ResolveExtensionDir(empty, baseDir, v0.2.0, cfg)
  -> installPath non-empty
  -> EmbedCompleteFS(installPath, extension) == true
  -> GETCount >= 1
```

## Preconditions

- Empty embed FS for extension tree.
- XDG temp; baseDir temp for optional materialize.
- Server serves `extension-complete` fixture as tar.gz.

## Steps

1. Set `ImplicitOp = ImplicitOpExtensionIncompleteEnsures`.
2. Embed empty; XDG + baseDir temp; serve extension-complete.

## Context

- Extension extract path when live embed incomplete (go install mini fixture).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ImplicitOp = ImplicitOpExtensionIncompleteEnsures
	req.ImplicitEmbedFixture = FixtureEmpty
	req.XDGCacheHome = t.TempDir()
	req.ImplicitBaseDir = t.TempDir()
	req.ImplicitVersion = CacheVersion
	req.ImplicitServeFixture = FixtureExtensionComplete
	return nil
}
```
