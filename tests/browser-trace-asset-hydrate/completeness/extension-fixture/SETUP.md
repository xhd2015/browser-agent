# Scenario

**Feature**: complete extension fixture is reported complete

```
testdata/extension-complete (manifest.json + background.js)
  -> EmbedCompleteFS(fs, "extension") -> true
```

## Preconditions

- Fixture has non-empty `manifest.json` and non-empty `background.js`.

## Steps

1. Set `FixtureName = FixtureExtensionComplete`.
2. Set `ExpectComplete = true`.
3. AssetKind remains extension (parent default).

## Context

- Happy path for extension completeness (shared EmbedCompleteFS rules).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FixtureName = FixtureExtensionComplete
	req.AssetKind = KindExtension
	req.ExpectComplete = true
	return nil
}
```
