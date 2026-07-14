## Expected

Requirement **C1**:

- HTTP 200 from `GET /v1/session`.
- `bundled_extension.version` and `.md5` present (or meta equivalents filled into resp).
- `extension_match` is `not_connected`.
- `extension.connected` is false.
- Preferred: stderr mentions embedded version and/or md5 (serve log).
- Preferred: meta.json has `extension_version` / `extension_md5`.

## Side Effects

- Serve extracts extension under BaseDir; writes meta + session artifacts.

## Errors

- Missing bundled identity or wrong match status fails the leaf.

## Exit Code

- Not asserted on process (in-process serve cancelled after probe).

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
	if match != "not_connected" {
		t.Fatalf("extension_match=%q, want not_connected; body=%s",
			match, truncate(resp.BodyString, 500))
	}
	if resp.ExtConnected {
		t.Fatal("extension.connected=true without hello")
	}

	// Soft prefer: serve logged embedded identity.
	combined := resp.Stderr + resp.Stdout
	if combined != "" {
		// Do not hard-fail if log wording differs; but if "embedded" appears,
		// require version or md5 nearby-ish via fold contains of known tokens.
		low := strings.ToLower(combined)
		if strings.Contains(low, "embedded") {
			if resp.BundledVersion != "" && !strings.Contains(combined, resp.BundledVersion) {
				// still soft — identity may be logged only as md5
				_ = resp.BundledVersion
			}
		}
	}

	// Meta preferred fields.
	if resp.Meta != nil {
		if resp.MetaExtVersion == "" && resp.MetaExtMD5 == "" {
			// Accept if only session JSON has bundled_extension (meta optional until implementer lands both).
			// Prefer non-empty when meta exists.
			t.Logf("meta.json present but missing extension_version/extension_md5; meta=%s",
				truncate(resp.MetaJSON, 300))
		}
	}
}
```
