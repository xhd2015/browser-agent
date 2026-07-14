## Expected

- `QueryDaemonStatus` returns **nil** error.
- `Running=true`.
- `ExtensionPath` is absolute and contains `extensions/browser-agent/` segment.
- `ExtensionVersion` and `ExtensionMD5` are non-empty.
- `server.json` bytes **unchanged** after query.

## Side Effects

- `EnsureCanonicalExtension` may extract under `TestHome` (idempotent).

## Errors

- Query error, empty extension fields, or meta mutation fails.

## Exit Code

- Not asserted.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.QueryErr != "" {
		t.Fatalf("QueryDaemonStatus error: %s", resp.QueryErr)
	}
	st := resp.Status
	if !st.Running {
		t.Fatal("Running=false, want true")
	}
	assertRichStatusFieldsPresent(t, st)
	extPath := richStatusString(st, "ExtensionPath")
	if !filepath.IsAbs(extPath) {
		t.Fatalf("ExtensionPath should be absolute; got %q", extPath)
	}
	if !strings.HasPrefix(filepath.ToSlash(extPath), filepath.ToSlash(req.TestHome)) {
		t.Fatalf("ExtensionPath %q should be under TestHome %q", extPath, req.TestHome)
	}
	assertMetaUnchanged(t, resp)
	if !resp.DaemonHealthyAfter {
		t.Fatal("daemon not healthy after read-only status query")
	}
}
```