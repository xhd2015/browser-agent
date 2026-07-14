## Expected

- `QueryDaemonStatus` returns **nil** error.
- `Running=false`.
- `DaemonVersion` empty or omitted (not running).
- `ExtensionPath` absolute with `extensions/browser-agent/` segment.
- `ExtensionVersion` and `ExtensionMD5` non-empty.
- `server.json` remains **absent**.

## Side Effects

- Canonical extension extract under `TestHome` is acceptable.

## Errors

- Query error, missing extension fields, or meta file created fails.

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
	if st.Running {
		t.Fatal("Running=true, want false without server.json")
	}
	if richStatusString(st, "DaemonVersion") != "" {
		t.Fatalf("DaemonVersion=%q want empty when not running", richStatusString(st, "DaemonVersion"))
	}
	extPath := richStatusString(st, "ExtensionPath")
	if extPath == "" {
		t.Fatal("ExtensionPath empty, want canonical path when not running")
	}
	assertCanonicalPathSegment(t, extPath)
	if !filepath.IsAbs(extPath) {
		t.Fatalf("ExtensionPath should be absolute; got %q", extPath)
	}
	if richStatusString(st, "ExtensionVersion") == "" {
		t.Fatal("ExtensionVersion empty")
	}
	if richStatusString(st, "ExtensionMD5") == "" {
		t.Fatal("ExtensionMD5 empty")
	}
	if resp.DaemonMetaBeforeHit || resp.DaemonMetaAfterHit {
		t.Fatal("server.json should remain absent for not-running query")
	}
	assertMetaUnchanged(t, resp)
}
```