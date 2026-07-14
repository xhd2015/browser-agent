## Expected

Requirement **C2**:

- HTTP 200.
- `bundled_extension` still present.
- `extension_match` is `ok`.
- `extension.connected` is true.
- Prefer: `extension.version` equals bundled version; `extension.bundle_md5` equals bundled md5 when present.
- No hard requirement on orange warning (should be absent).

## Side Effects

- Fake WS stays connected until cleanup.

## Errors

- match≠ok or connected false fails the leaf.

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
	assertHTTP200(t, resp)
	assertBundledExtensionPresent(t, resp)

	match := strings.TrimSpace(resp.ExtensionMatch)
	if match != "ok" {
		t.Fatalf("extension_match=%q, want ok; body=%s stderr=%s",
			match, truncate(resp.BodyString, 500), truncate(resp.Stderr, 400))
	}
	if !resp.ExtConnected {
		t.Fatalf("extension.connected=false after matching hello; body=%s",
			truncate(resp.BodyString, 400))
	}
	if resp.BundledVersion != "" && resp.ExtVersion != "" && resp.ExtVersion != resp.BundledVersion {
		t.Fatalf("extension.version=%q != bundled %q", resp.ExtVersion, resp.BundledVersion)
	}
	if resp.BundledMD5 != "" && resp.ExtBundleMD5 != "" {
		if strings.ToLower(resp.ExtBundleMD5) != strings.ToLower(resp.BundledMD5) {
			t.Fatalf("extension.bundle_md5=%q != bundled %q", resp.ExtBundleMD5, resp.BundledMD5)
		}
	}
}
```
