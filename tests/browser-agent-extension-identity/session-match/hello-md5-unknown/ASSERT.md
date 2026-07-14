## Expected

Requirement **C5**:

- HTTP 200; connected true.
- `extension_match` is `md5_unknown`.
- `bundled_extension` present.
- Prefer: warning logged for md5_unknown (soft).

## Side Effects

- Session usable without hard-fail.

## Errors

- match≠md5_unknown fails the leaf.

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
	if match != "md5_unknown" {
		t.Fatalf("extension_match=%q, want md5_unknown; body=%s stderr=%s",
			match, truncate(resp.BodyString, 500), truncate(resp.Stderr, 500))
	}
	if !resp.ExtConnected {
		t.Fatal("extension.connected=false after hello without md5")
	}
}
```
