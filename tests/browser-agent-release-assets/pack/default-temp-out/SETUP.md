# Scenario

**Feature**: omit `--out` → script creates a temp dir, packs three archives, exits 0

```
go run ./script/github/release-assets --version v0.2.0
  # (no --out)
  -> creates temp dir (MkdirTemp browser-agent-release-assets-*)
  -> writes 3 .tar.gz with AssetReleaseNames basenames
  -> prints out: <abs-path> on stdout (preferred)
  exit 0
```

## Preconditions

- Parent pack Setup sets ModePack and defaults Version.
- Parent may allocate OutDir / Args with `--out`; this leaf **overrides** to omit `--out`.
- Embed sources exist under ModuleRoot.

## Steps

1. Set `PackOmitOut = true`.
2. Pin `Version = ReleaseVersion` (`v0.2.0`).
3. Clear `OutDir` so Run discovers the script-created path from stdout.
4. Set Args to **only** `--version` + version (no `--out`).

## Context

- Classic TDD: today the script fails with `--out DIR is required for pack` — expect **RED** until implementer defaults `--out` to a temp dir.
- Prefer stdout token `out: <abs-path>`; Run also accepts a parseable “packing into …” path.
- Do not require the script to delete the temp dir on success.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	if req.Mode != ModePack {
		t.Fatalf("Mode=%q want %q", req.Mode, ModePack)
	}
	req.PackOmitOut = true
	req.Version = ReleaseVersion
	// Force script default temp out — do not pass --out.
	req.OutDir = ""
	req.Args = []string{
		"--version", req.Version,
	}
	return nil
}
```
