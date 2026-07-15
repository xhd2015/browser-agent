# Scenario

**Feature**: implicit hydrate — embed first, else EnsureAsset + cache (P4)

```
ResolveSessionPage(embedFS, version, cfg)
  complete -> source=embed, no GET
  incomplete -> EnsureAsset session-page -> source=cache

ResolveExtensionDir(embedFS, baseDir, version, cfg)
  incomplete -> EnsureAsset extension -> complete installPath
```

## Preconditions

- Mode is implicit.
- Leaf sets ImplicitOp, embed fixture, XDG temp, optional httptest.

## Steps

1. Set `Mode = ModeImplicit`.
2. Default version `v0.2.0`.

## Context

- Classic TDD for P4 — RED until ResolveSessionPage / ResolveExtensionDir exist.
- Do not mutate live //go:embed; use injectable fixture FS.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeImplicit
	if req.ImplicitVersion == "" {
		req.ImplicitVersion = CacheVersion
	}
	return nil
}
```
