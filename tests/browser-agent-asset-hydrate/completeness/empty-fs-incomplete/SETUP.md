# Scenario

**Feature**: empty FS is incomplete for both asset kinds

```
empty temp DirFS
  -> EmbedCompleteFS(fs, "session-page") -> false
  -> EmbedCompleteFS(fs, "extension") -> false
```

## Preconditions

- No files under the FS root (missing keys / empty tree).
- AssetKind left empty so Run checks both kinds.

## Steps

1. Set `FixtureName = FixtureEmpty` (synthetic empty temp dir).
2. Clear `AssetKind` (empty string) so Run inspects both kinds.
3. Set `ExpectComplete = false`.

## Context

- Covers missing-key / empty embed trees for both products kinds in one leaf.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FixtureName = FixtureEmpty
	req.AssetKind = "" // both kinds
	req.ExpectComplete = false
	return nil
}
```
