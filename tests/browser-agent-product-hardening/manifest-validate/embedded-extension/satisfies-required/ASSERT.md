## Expected

Requirement **A4**:

- Embedded `manifest.json` is readable (ManifestText non-empty).
- `ValidateExtensionManifestJSON` returns nil (`ValidateOK` true).
- Manifest text includes required permission names and **43761** (defense in
  depth so a no-op validator cannot hide a stripped production file).

## Side Effects

- None.

## Errors

- Missing file is a Run transport failure; soft validate-ok with empty text fails.

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
		t.Fatalf("embedded manifest empty; path=%q", resp.ManifestPath)
	}
	if !resp.ValidateOK {
		t.Fatalf("embedded manifest must validate; ValidateErr=%q path=%q",
			resp.ValidateErr, resp.ManifestPath)
	}
	text := resp.ManifestText
	for _, need := range []string{"debugger", "tabs", "alarms", "storage", "43761"} {
		if !strings.Contains(text, need) {
			t.Fatalf("embedded manifest must contain %q; path=%s text=%s",
				need, resp.ManifestPath, truncate(text, 500))
		}
	}
	// Broad host: <all_urls> preferred; accept explicit all_urls token.
	if !strings.Contains(text, "<all_urls>") && !strings.Contains(text, "*://*/*") {
		t.Fatalf("embedded manifest must include broad host access (<all_urls> or equiv); path=%s",
			resp.ManifestPath)
	}
	assertExitZero(t, resp)
}
```
