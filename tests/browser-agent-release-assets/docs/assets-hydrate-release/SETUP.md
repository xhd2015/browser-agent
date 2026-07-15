# Scenario

**Feature**: docs/assets-hydrate.md mentions script/github/release-assets and --upload

```
docs/assets-hydrate.md
  contains:
    script/github/release-assets   (or go run ./script/github/release-assets)
    --upload
```

## Preconditions

- DocsOp = assets-hydrate-release.
- File may exist today without release-assets section → leaf RED until doc update.

## Steps

1. Set `DocsOp = DocsOpAssetsHydrateRelease`.

## Context

- Pins preferred path `docs/assets-hydrate.md` (not README/SKILL fallback).
- Implementer should document pack + optional `--upload` (gh create / --clobber).
- Prefer real path tokens; no dotted scaffold placeholder IDs.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DocsOp = DocsOpAssetsHydrateRelease
	return nil
}
```
