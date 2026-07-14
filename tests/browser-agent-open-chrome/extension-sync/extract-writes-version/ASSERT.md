## Expected

- `EnsureManagedExtension` returns nil error.
- `ExtensionPath` is absolute; base name equals `ExtensionVer`.
- `ExtensionPath` is under `Layout.ExtensionsDir`.
- `manifest.json` exists at `ExtensionPath/manifest.json`.
- Manifest JSON contains `"version"` matching `ExtensionVer`.

## Side Effects

- Writes extension tree under managed `extensions/{version}/`.

## Errors

- Missing manifest or empty version fails.

## Exit Code

- N/A.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("EnsureManagedExtension error: %v", err)
	}
	if resp.ExtensionVer == "" {
		t.Fatal("ExtensionVer is empty")
	}
	if !filepath.IsAbs(resp.ExtensionPath) {
		t.Fatalf("ExtensionPath %q is not absolute", resp.ExtensionPath)
	}
	if filepath.Base(resp.ExtensionPath) != resp.ExtensionVer {
		t.Fatalf("ExtensionPath base %q != version %q", filepath.Base(resp.ExtensionPath), resp.ExtensionVer)
	}
	rel, err := filepath.Rel(resp.Layout.ExtensionsDir, resp.ExtensionPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("ExtensionPath %q not under ExtensionsDir %q (rel=%q)", resp.ExtensionPath, resp.Layout.ExtensionsDir, rel)
	}
	data, err := os.ReadFile(resp.ManifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if !strings.Contains(string(data), "version") {
		t.Fatalf("manifest.json missing version field; data=%s", truncate(string(data), 200))
	}
	if !strings.Contains(string(data), resp.ExtensionVer) {
		t.Fatalf("manifest.json should mention version %q", resp.ExtensionVer)
	}
}```
