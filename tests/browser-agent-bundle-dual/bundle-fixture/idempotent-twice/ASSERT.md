## Expected

Requirement **A3**:

- First and second Bundle both succeed (nil error, exit 0).
- `SecondExtensionDir` equals `ExtensionDir` (cleaned path compare OK).
- `SecondSessionPageDir` equals `SessionPageDir`.
- Both dirs still exist and contain expected files (manifest / index).

## Side Effects

- Still confined to BundleRoot; re-stage may rewrite file contents but paths stable.

## Errors

- Path drift between passes breaks embed/CI repeatability.

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
	if req.BundlePasses < 2 {
		t.Fatalf("BundlePasses = %d, want >= 2", req.BundlePasses)
	}
	if strings.TrimSpace(resp.ExtensionDir) == "" || strings.TrimSpace(resp.SessionPageDir) == "" {
		t.Fatalf("first pass paths empty: ext=%q session=%q",
			resp.ExtensionDir, resp.SessionPageDir)
	}
	if strings.TrimSpace(resp.SecondExtensionDir) == "" || strings.TrimSpace(resp.SecondSessionPageDir) == "" {
		t.Fatalf("second pass paths empty: ext=%q session=%q",
			resp.SecondExtensionDir, resp.SecondSessionPageDir)
	}
	ext1 := filepath.Clean(resp.ExtensionDir)
	ext2 := filepath.Clean(resp.SecondExtensionDir)
	if ext1 != ext2 {
		t.Fatalf("ExtensionDir unstable: first=%q second=%q", ext1, ext2)
	}
	sp1 := filepath.Clean(resp.SessionPageDir)
	sp2 := filepath.Clean(resp.SecondSessionPageDir)
	if sp1 != sp2 {
		t.Fatalf("SessionPageDir unstable: first=%q second=%q", sp1, sp2)
	}
	if _, err := os.Stat(filepath.Join(ext1, "manifest.json")); err != nil {
		t.Fatalf("manifest missing after second Bundle: %v", err)
	}
	foundIndex := false
	for _, name := range []string{"index.html", "session-page.html"} {
		if _, err := os.Stat(filepath.Join(sp1, name)); err == nil {
			foundIndex = true
			break
		}
	}
	if !foundIndex {
		t.Fatalf("session index missing after second Bundle under %s", sp1)
	}
}
```
