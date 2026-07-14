## Expected

- `LayoutFromRoot(ManagedRoot)` returns layout with absolute `Root`.
- `DataDir` and `ExtensionsDir` are children of `Root`.

## Side Effects

- None.

## Errors

- Non-absolute root or wrong child paths fails.

## Exit Code

- N/A.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("LayoutFromRoot error: %v", err)
	}
	if !filepath.IsAbs(resp.Layout.Root) {
		t.Fatalf("Layout.Root %q is not absolute", resp.Layout.Root)
	}
	assertLayoutUnderRoot(t, resp.Layout, resp.Layout.Root)
}```
