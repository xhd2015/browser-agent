# Scenario

**Feature**: EmbedCompleteFS reports extension completeness for browser-trace fixtures

```
# pure completeness against injectable fs.FS (extension kind only)
Test Client -> openFixtureFS(fixture)
  -> browseragent.EmbedCompleteFS(fs, "extension")
  -> bool
```

## Preconditions

- Mode is completeness.
- Leaf sets FixtureName (and optionally AssetKind; default extension).
- FS rooted at extension tree (`manifest.json` + `background.js` at DirFS root).

## Steps

1. Set `Mode = ModeCompleteness`.
2. Default AssetKind to extension when leaf does not set it.

## Context

- browser-trace product has no session-page embed — only extension completeness.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCompleteness
	if req.AssetKind == "" {
		req.AssetKind = KindExtension
	}
	return nil
}
```
