## Expected

- `ResolveExtensionDir` err nil.
- `InstallPath` non-empty.
- `ExtComplete` true (`EmbedCompleteFS` on install path for extension:
  non-empty `manifest.json` + `background.js`).
- `GETCount >= 1`.
- Prefer install path under XDG cache or under leaf `ImplicitBaseDir`.

## Side Effects

- Complete extension tree on disk under cache and/or baseDir (isolated temps).

## Errors

- Install error, empty path, incomplete tree, or no GET fails.

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.InstallErr != nil {
		t.Fatalf("ResolveExtensionDir err=%v", resp.InstallErr)
	}
	if strings.TrimSpace(resp.InstallPath) == "" {
		t.Fatal("InstallPath empty")
	}
	// Path should be under XDG and/or baseDir isolation roots.
	underXDG := false
	underBase := false
	if req.XDGCacheHome != "" {
		if rel, rerr := filepath.Rel(filepath.Clean(req.XDGCacheHome), filepath.Clean(resp.InstallPath)); rerr == nil && !strings.HasPrefix(rel, "..") {
			underXDG = true
		}
	}
	if req.ImplicitBaseDir != "" {
		if rel, rerr := filepath.Rel(filepath.Clean(req.ImplicitBaseDir), filepath.Clean(resp.InstallPath)); rerr == nil && !strings.HasPrefix(rel, "..") {
			underBase = true
		}
	}
	if !underXDG && !underBase {
		t.Fatalf("InstallPath=%q not under XDG %q or baseDir %q",
			resp.InstallPath, req.XDGCacheHome, req.ImplicitBaseDir)
	}
	if !resp.ExtComplete {
		t.Fatalf("install path %q is not extension-complete (manifest+background)", resp.InstallPath)
	}
	if resp.GETCount < 1 {
		t.Fatalf("GETCount=%d want >= 1", resp.GETCount)
	}
	assertExitZero(t, resp)
}
```
