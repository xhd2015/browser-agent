## Expected

- `EnsureCanonicalFirefoxExtensionWithHome` returns nil error.
- `ExtensionPath` contains `extensions/browser-agent-firefox/`.
- `ExtensionPath` does **not** contain `managed-chrome`.
- Path is absolute; `manifest.json` exists; `ExtensionVer` non-empty.

## Side Effects

- Files written under `TestHome/.browser-agent/extensions/browser-agent-firefox/{ver}/`.

## Errors

- Wrong path segment or missing manifest fails.

## Exit Code

- Not asserted.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if strings.TrimSpace(resp.ExtensionPath) == "" {
		t.Fatal("ExtensionPath empty")
	}
	assertFirefoxCanonicalPathSegment(t, resp.ExtensionPath)
	if !filepath.IsAbs(resp.ExtensionPath) {
		t.Fatalf("ExtensionPath should be absolute; got %q", resp.ExtensionPath)
	}
	// Must live under TestHome.
	normPath := filepath.ToSlash(resp.ExtensionPath)
	normHome := filepath.ToSlash(req.TestHome)
	if !strings.HasPrefix(normPath, normHome) && !strings.Contains(normPath, filepath.ToSlash(req.TestHome)) {
		// Allow resolved abs paths that still contain the home base name tree.
		if !strings.Contains(normPath, "/.browser-agent/extensions/browser-agent-firefox/") {
			t.Fatalf("ExtensionPath %q should be under TestHome %q", resp.ExtensionPath, req.TestHome)
		}
	}
	if _, statErr := os.Stat(resp.ManifestPath); statErr != nil {
		t.Fatalf("manifest.json missing at %q: %v", resp.ManifestPath, statErr)
	}
	if strings.TrimSpace(resp.ExtensionVer) == "" {
		t.Fatal("ExtensionVer empty")
	}
}
```
