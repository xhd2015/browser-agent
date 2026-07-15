# Scenario

**Feature**: empty FS is incomplete for extension kind

```
empty temp DirFS
  -> EmbedCompleteFS(fs, "extension") -> false
```

## Preconditions

- No files under the FS root (missing keys / empty tree).

## Steps

1. Set `FixtureName = FixtureEmpty` (synthetic empty temp dir).
2. Set `AssetKind = KindExtension`.
3. Set `ExpectComplete = false`.

## Context

- Covers missing-key / empty embed trees for browser-trace extension only.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FixtureName = FixtureEmpty
	req.AssetKind = KindExtension
	req.ExpectComplete = false
	return nil
}
```
