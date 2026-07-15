# Scenario

**Feature**: operator docs cross-link release-assets pack + upload (P7)

```
Test Client -> read docs/assets-hydrate.md under ModuleRoot
  -> documents go run ./script/github/release-assets pack
  -> documents --upload (gh create / clobber)
```

## Preconditions

- Mode is docs.
- ModuleRoot resolved by root Setup.
- Preferred path: `docs/assets-hydrate.md` (AssetsHydrateDocRel).

## Steps

1. Set `Mode = ModeDocs`.

## Context

- Classic TDD — RED until implementer updates `docs/assets-hydrate.md`.
- FS-only; no network; no script execution.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeDocs
	return nil
}
```
