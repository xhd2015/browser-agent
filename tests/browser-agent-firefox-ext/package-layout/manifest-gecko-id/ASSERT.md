## Expected

- `Firefox-Ext-Browser-Agent/public/manifest.json` exists under ModuleRoot.
- `manifest_version` is 3.
- `browser_specific_settings.gecko.id` is a non-empty string.

## Side Effects

- None (read-only).

## Errors

- Missing package, missing gecko.id, or non-MV3 fails.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.ManifestText) == "" {
		t.Fatalf("Firefox-Ext-Browser-Agent/public/manifest.json missing under ModuleRoot=%s path=%s",
			req.ModuleRoot, resp.ManifestPath)
	}
	if resp.ManifestVersion != 3 {
		t.Fatalf("manifest_version = %d, want 3; path=%s", resp.ManifestVersion, resp.ManifestPath)
	}
	if strings.TrimSpace(resp.GeckoID) == "" {
		t.Fatalf("browser_specific_settings.gecko.id must be non-empty; path=%s manifest=%s",
			resp.ManifestPath, truncate(resp.ManifestText, 400))
	}
}
```
