## Expected

- Manifest exists under `Firefox-Ext-Browser-Agent/public/`.
- Permissions array does **not** include `"debugger"`.
- `HasDebuggerPerm` is false.

## Side Effects

- None (read-only).

## Errors

- Missing manifest or debugger permission present fails.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.ManifestText) == "" {
		t.Fatalf("Firefox public/manifest.json missing under ModuleRoot=%s path=%s",
			req.ModuleRoot, resp.ManifestPath)
	}
	if resp.HasDebuggerPerm {
		t.Fatalf("Firefox package must NOT include debugger permission; permissions=%v path=%s",
			resp.Permissions, resp.ManifestPath)
	}
	// Defense in depth: raw text must not list "debugger" as a permission token.
	// Allow the word only if absent from permissions array (already checked).
	for _, p := range resp.Permissions {
		if p == "debugger" {
			t.Fatalf("permissions contains debugger: %v", resp.Permissions)
		}
	}
}
```
