# Scenario

**Feature**: EmbedCompleteFS reports whether outstanding embed files exist

```
# pure completeness against injectable fs.FS
Test Client -> openFixtureFS(fixture)
  -> browseragent.EmbedCompleteFS(fs, kind)
  -> bool
```

## Preconditions

- Mode is completeness.
- Leaf sets AssetKind and/or FixtureName.
- FS rooted at asset tree (index.html / manifest.json at root of DirFS).

## Steps

1. Set `Mode = ModeCompleteness`.

## Context

- Kinds: `session-page`, `extension`.
- Empty AssetKind means both kinds checked (empty-fs leaf).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCompleteness
	return nil
}
```
