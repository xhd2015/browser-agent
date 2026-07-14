## Expected

Requirement **D2**:

- `Chrome-Ext-Capture-API` exists under ModuleRoot.
- `manifest.json` found.
- Manifest text contains **43759**.
- Prefer identity markers: "API Capture" and/or "browser-trace" (soft: at least one).

## Side Effects

- None.

## Errors

- Manifest on 43761 would collide with browser-agent.

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
		t.Fatalf("Chrome-Ext-Capture-API manifest missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	if !strings.Contains(resp.ManifestText, "43759") {
		t.Fatalf("capture manifest must mention port 43759; path=%s manifest=%s",
			resp.ManifestPath, truncate(resp.ManifestText, 400))
	}
	// Soft identity: API Capture or browser-trace
	low := strings.ToLower(resp.ManifestText)
	if !strings.Contains(low, "api capture") &&
		!strings.Contains(low, "browser-trace") &&
		!strings.Contains(low, "browser trace") &&
		!strings.Contains(low, "har") {
		t.Fatalf("capture manifest should reference API Capture / browser-trace / HAR; path=%s manifest=%s",
			resp.ManifestPath, truncate(resp.ManifestText, 400))
	}
	// Hard: must not be agent-only shell
	if strings.Contains(resp.ManifestText, "43761") && !strings.Contains(resp.ManifestText, "43759") {
		t.Fatal("capture manifest has 43761 without 43759")
	}
}
```
