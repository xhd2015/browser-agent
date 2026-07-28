## Expected

- First and second calls return nil error.
- `ExtensionPath2 == ExtensionPath`.
- `ExtensionVer2 == ExtensionVer`.
- Manifest still present after second call.
- Path remains under `extensions/browser-agent-firefox/` (not managed-chrome).

## Side Effects

- No duplicate version directories required.

## Errors

- Path or version drift on second call fails.

## Exit Code

- Not asserted.

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExtensionPath2 != resp.ExtensionPath {
		t.Fatalf("second path %q != first %q", resp.ExtensionPath2, resp.ExtensionPath)
	}
	if resp.ExtensionVer2 != resp.ExtensionVer {
		t.Fatalf("second version %q != first %q", resp.ExtensionVer2, resp.ExtensionVer)
	}
	assertFirefoxCanonicalPathSegment(t, resp.ExtensionPath)
	if _, statErr := os.Stat(resp.ManifestPath); statErr != nil {
		t.Fatalf("manifest missing after second call: %v", statErr)
	}
}
```
