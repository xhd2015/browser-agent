## Expected

Requirement **C4**:

- HTTP 200; connected true.
- `extension_match` is `md5_mismatch`.
- `bundled_extension` present.
- Prefer: stderr mentions mismatch (soft if only JSON status is set).

## Side Effects

- Warning may be logged; jobs not hard-failed (not asserted).

## Errors

- match≠md5_mismatch fails the leaf.

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
	if match != "md5_mismatch" {
		t.Fatalf("extension_match=%q, want md5_mismatch; body=%s stderr=%s",
			match, truncate(resp.BodyString, 500), truncate(resp.Stderr, 500))
	}
	if !resp.ExtConnected {
		t.Fatal("extension.connected=false after hello")
	}
	// Prefer loaded md5 reflected on extension snapshot.
	if resp.ExtBundleMD5 != "" {
		got := strings.ToLower(resp.ExtBundleMD5)
		want := strings.ToLower(req.ForceHelloMD5)
		if want != "" && got != want {
			t.Fatalf("extension.bundle_md5=%q, want hello md5 %q", resp.ExtBundleMD5, req.ForceHelloMD5)
		}
	}
}
```
