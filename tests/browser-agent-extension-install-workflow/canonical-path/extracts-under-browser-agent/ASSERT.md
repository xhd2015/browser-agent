## Expected

- `EnsureCanonicalExtension` returns nil error.
- `ExtensionPath` contains `extensions/browser-agent/` segment.
- `manifest.json` exists at `ManifestPath`.
- `ExtensionVer` is non-empty.

## Side Effects

- Files written under `TestHome/.browser-agent/managed-chrome/extensions/browser-agent/{ver}/`.

## Errors

- Missing manifest or wrong path segment fails.

## Exit Code

- Not asserted.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if strings.TrimSpace(resp.ExtensionPath) == "" {
		t.Fatal("ExtensionPath empty")
	}
	assertCanonicalPathSegment(t, resp.ExtensionPath)
	if !filepath.IsAbs(resp.ExtensionPath) {
		t.Fatalf("ExtensionPath should be absolute; got %q", resp.ExtensionPath)
	}
	if _, statErr := os.Stat(resp.ManifestPath); statErr != nil {
		t.Fatalf("manifest.json missing at %q: %v", resp.ManifestPath, statErr)
	}
	if strings.TrimSpace(resp.ExtensionVer) == "" {
		t.Fatal("ExtensionVer empty")
	}
}
```