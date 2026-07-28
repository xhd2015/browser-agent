## Expected

- `BuildFirefoxExtensionShell` returns a non-nil product error (`BuildShellErr` non-empty).
- `BuildDir` is empty (no successful stage).

## Side Effects

- None required on disk.

## Errors

- Success without error fails this leaf (must error on missing public).

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	// Run swallows product error into BuildShellErr for this op.
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if strings.TrimSpace(resp.BuildShellErr) == "" {
		t.Fatalf("expected error when public/manifest.json missing under %s; BuildDir=%q",
			req.ShellRoot, resp.BuildDir)
	}
	// Error message should mention missing public/manifest (flexible wording).
	assertContainsFold(t, resp.BuildShellErr, "manifest")
}
```
