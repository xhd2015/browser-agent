## Expected

Requirement **A4**:

- Bundle succeeds.
- Staged extension content (manifest and/or scripts under ExtensionDir)
  mentions **43761**.
- Should not be only-43759 agent fixture (hard fail if 43761 absent).

## Side Effects

- None beyond Bundle stage under BundleRoot.

## Errors

- Staging a browser-trace (43759-only) fixture breaks dual-product split.

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
	text := resp.ExtensionCombinedText
	if text == "" {
		text = resp.ManifestText
	}
	if text == "" && resp.ExtensionDir != "" {
		// last-chance read manifest
		if b, e := os.ReadFile(filepath.Join(resp.ExtensionDir, "manifest.json")); e == nil {
			text = string(b)
		}
	}
	if !strings.Contains(text, "43761") {
		t.Fatalf("staged extension must mention port 43761; dir=%s text=%s",
			resp.ExtensionDir, truncate(text, 500))
	}
}
```
