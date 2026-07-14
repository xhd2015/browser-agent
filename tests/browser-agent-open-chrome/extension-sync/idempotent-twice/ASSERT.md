## Expected

- First and second calls return nil error.
- `ExtensionPath2 == ExtensionPath`.
- `ExtensionVer2 == ExtensionVer`.
- Manifest still present after second call.

## Side Effects

- No duplicate version directories.

## Errors

- Path or version drift on second call fails.

## Exit Code

- N/A.

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("EnsureManagedExtension error: %v", err)
	}
	if resp.ExtensionPath2 != resp.ExtensionPath {
		t.Fatalf("second path %q != first %q", resp.ExtensionPath2, resp.ExtensionPath)
	}
	if resp.ExtensionVer2 != resp.ExtensionVer {
		t.Fatalf("second version %q != first %q", resp.ExtensionVer2, resp.ExtensionVer)
	}
	if _, err := os.Stat(resp.ManifestPath); err != nil {
		t.Fatalf("manifest missing after second call: %v", err)
	}
}```
