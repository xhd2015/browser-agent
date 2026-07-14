## Expected

Requirement **A5**:

- Shell `public/manifest.json` is readable (ManifestText non-empty).
- `ValidateExtensionManifestJSON` returns nil (`ValidateOK` true).
- Manifest text includes required permission names and **43761**.
- Broad host access present (`<all_urls>` or equivalent).

## Side Effects

- None.

## Errors

- Pointing only at Capture-API / 43759 without agent perms fails product split.

## Exit Code

- 0.

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
	if strings.TrimSpace(resp.ManifestText) == "" {
		t.Fatalf("shell manifest empty; path=%q ModuleRoot=%s",
			resp.ManifestPath, req.ModuleRoot)
	}
	if !resp.ValidateOK {
		t.Fatalf("shell public manifest must validate; ValidateErr=%q path=%q",
			resp.ValidateErr, resp.ManifestPath)
	}
	text := resp.ManifestText
	for _, need := range []string{"debugger", "tabs", "alarms", "storage", "43761"} {
		if !strings.Contains(text, need) {
			t.Fatalf("shell manifest must contain %q; path=%s text=%s",
				need, resp.ManifestPath, truncate(text, 500))
		}
	}
	if !strings.Contains(text, "<all_urls>") && !strings.Contains(text, "*://*/*") {
		t.Fatalf("shell manifest must include broad host access (<all_urls> or equiv); path=%s",
			resp.ManifestPath)
	}
	assertExitZero(t, resp)
}
```
