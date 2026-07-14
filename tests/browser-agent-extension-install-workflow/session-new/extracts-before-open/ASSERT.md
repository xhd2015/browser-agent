## Expected

- `SessionNew` succeeds.
- `ExtensionPath` non-empty with `extensions/browser-agent/` segment.
- `manifest.json` exists under extension path.

## Side Effects

- Canonical dir created before chrome launch.

## Errors

- Missing extract or legacy managed-only path fails.

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
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNew error: %s", resp.SessionNewErr)
	}
	if strings.TrimSpace(resp.ExtensionPath) == "" {
		t.Fatal("ExtensionPath empty after SessionNew")
	}
	assertCanonicalPathSegment(t, resp.ExtensionPath)
	manifest := filepath.Join(resp.ExtensionPath, "manifest.json")
	if _, statErr := os.Stat(manifest); statErr != nil {
		t.Fatalf("manifest missing at %q: %v", manifest, statErr)
	}
}
```