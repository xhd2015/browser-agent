## Expected

Requirement **A1**:

- `Bundle(UseFixture)` succeeds (nil error, exit 0).
- `ExtensionDir` is non-empty absolute path under BundleRoot.
- `ExtensionDir/manifest.json` exists and is readable.
- Manifest `"version"` is non-empty (Response.ManifestVersion or JSON parse).
- `UsedFixture` is true (when field populated).

## Side Effects

- Writes only under BundleRoot (temp); not under live ModuleRoot embed.

## Errors

- Missing manifest or empty version fails the fixture contract.

## Exit Code

- 0.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertExitZero(t, resp)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if strings.TrimSpace(resp.ExtensionDir) == "" {
		t.Fatal("ExtensionDir is empty")
	}
	if !filepath.IsAbs(resp.ExtensionDir) {
		t.Fatalf("ExtensionDir must be absolute; got %q", resp.ExtensionDir)
	}
	// Prefer path under BundleRoot when BundleRoot set
	if req.BundleRoot != "" {
		rel, rerr := filepath.Rel(req.BundleRoot, resp.ExtensionDir)
		if rerr != nil || strings.HasPrefix(rel, "..") {
			t.Fatalf("ExtensionDir %q not under BundleRoot %q", resp.ExtensionDir, req.BundleRoot)
		}
	}
	mp := resp.ManifestPath
	if mp == "" {
		mp = filepath.Join(resp.ExtensionDir, "manifest.json")
	}
	st, err := os.Stat(mp)
	if err != nil || st.IsDir() {
		t.Fatalf("manifest.json missing at %s: %v", mp, err)
	}
	ver := strings.TrimSpace(resp.ManifestVersion)
	if ver == "" {
		// try parse from ManifestText
		ver = strings.TrimSpace(parseManifestVersion(resp.ManifestText))
	}
	if ver == "" {
		t.Fatalf("manifest version is empty; path=%s text=%s", mp, truncate(resp.ManifestText, 300))
	}
	if !resp.UsedFixture {
		// Soft preference: UsedFixture should be true for fixture path.
		// Fail hard — requirement says fixture pipeline.
		t.Fatal("UsedFixture = false, want true for UseFixture Bundle")
	}
}
```
