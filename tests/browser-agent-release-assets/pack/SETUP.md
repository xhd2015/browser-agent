# Scenario

**Feature**: pack-only mode writes release asset archives under --out (or temp default)

```
go run ./script/github/release-assets --out DIR --version v0.2.0
  # (no --upload)
  -> three .tar.gz from embeds
  -> basenames = AssetReleaseNames(v0.2.0)

# default-temp-out leaf overrides: omit --out (PackOmitOut)
go run ./script/github/release-assets --version v0.2.0
```

## Preconditions

- Mode is pack.
- Leaf allocates temp OutDir and sets Version when needed (explicit `--out` leaves).
- `default-temp-out` clears OutDir / sets PackOmitOut after this Setup.
- Working directory for `go run` is ModuleRoot.

## Steps

1. Set `Mode = ModePack`.
2. Allocate temp `OutDir` when empty (explicit-out default for sibling leaves).
3. Default `Version` to `ReleaseVersion` (`v0.2.0`) when empty.
4. Default Args to `--out` + `--version` when empty (overridden by PackOmitOut leaf).

## Context

- Pack must not call `gh` or require network.
- Embed dirs are the live module trees (not fixtures).
- `PackOmitOut` leaves must re-set `Args` without `--out` and clear `OutDir`.

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModePack
	if req.OutDir == "" {
		req.OutDir = t.TempDir()
	}
	if err := os.MkdirAll(req.OutDir, 0o755); err != nil {
		return err
	}
	if req.Version == "" {
		req.Version = ReleaseVersion
	}
	// Explicit args so implementer flags are pinned by leaves that care;
	// default Run path also builds --out/--version when Args empty.
	// PackOmitOut leaf overwrites Args and clears OutDir after this Setup.
	if len(req.Args) == 0 {
		req.Args = []string{
			"--out", req.OutDir,
			"--version", req.Version,
		}
	}
	return nil
}
```
