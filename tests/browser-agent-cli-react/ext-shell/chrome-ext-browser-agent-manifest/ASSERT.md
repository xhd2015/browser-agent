## Expected

Requirement **H1**:

- Directory `Chrome-Ext-Browser-Agent` exists under ModuleRoot.
- manifest.json found (public/ or root or src/ or build/).
- Manifest text references Browser Agent (name or description; case-insensitive
  "browser agent" OK).
- Manifest text contains **43761**.

## Side Effects

- None.

## Errors

- Pointing only at Capture-API / 43759 fails product split.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.ManifestText) == "" {
		t.Fatalf("Chrome-Ext-Browser-Agent manifest missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	low := strings.ToLower(resp.ManifestText)
	if !strings.Contains(low, "browser agent") && !strings.Contains(low, "browser-agent") {
		t.Fatalf("manifest name/description should reference Browser Agent; manifest=%s",
			truncate(resp.ManifestText, 400))
	}
	if !strings.Contains(resp.ManifestText, "43761") {
		t.Fatalf("manifest must mention host port 43761; manifest=%s",
			truncate(resp.ManifestText, 400))
	}
}
```
