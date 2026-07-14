## Expected

Requirement **C3**:

- HTTP 200; extension connected true.
- `extension_match` is `version_mismatch`.
- Warning text (stderr or extracted) mentions both embedded and loaded versions
  when available (fold-contains version strings / "mismatch").
- `bundled_extension` still present.

## Side Effects

- Warning logged; session remains usable (v1 no job hard-fail).

## Errors

- match≠version_mismatch fails the leaf.

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
	if match != "version_mismatch" {
		t.Fatalf("extension_match=%q, want version_mismatch; body=%s stderr=%s",
			match, truncate(resp.BodyString, 500), truncate(resp.Stderr, 500))
	}
	if !resp.ExtConnected {
		t.Fatal("extension.connected=false after hello (mismatch must still connect)")
	}

	warn := resp.WarningText
	if warn == "" {
		warn = resp.Stderr
	}
	// Must look like a mismatch warning and mention both sides when known.
	assertContainsFold(t, warn, "mismatch")
	loaded := req.ForceHelloVersion
	if loaded == "" {
		loaded = "9.9.9"
	}
	if !strings.Contains(warn, loaded) {
		// Also accept extension.version echoed only in session JSON, but prefer stderr.
		if resp.ExtVersion != loaded {
			t.Fatalf("warning/stderr should mention loaded version %q; warn=%s stderr=%s",
				loaded, truncate(warn, 400), truncate(resp.Stderr, 400))
		}
	}
	if resp.BundledVersion != "" && !strings.Contains(warn, resp.BundledVersion) {
		// Prefer embedded version in warning; if implementer only logs match enum, require mismatch token (already checked).
		t.Logf("warning missing embedded version %q; warn=%s", resp.BundledVersion, truncate(warn, 300))
	}
}
```
