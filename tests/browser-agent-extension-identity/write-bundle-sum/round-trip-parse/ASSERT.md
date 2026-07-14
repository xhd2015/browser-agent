## Expected

Requirement **B3**:

- Write succeeds (`ParseOK` after re-parse).
- Parsed `Version` equals WriteVersion.
- Parsed `MD5` equals WriteMD5 (case-insensitive hex).
- Written file contains both assignment tokens.
- ExitCode 0.

## Side Effects

- Creates `bundle-sum.js` under temp extension dir.

## Errors

- Write or parse failure fails the leaf.

## Exit Code

- 0.

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
	if !resp.ParseOK {
		t.Fatalf("round-trip failed; ParseErr=%q ErrText=%q written=%s",
			resp.ParseErr, resp.ErrText, truncate(resp.WrittenJS, 300))
	}
	if resp.Version != req.WriteVersion {
		t.Fatalf("Version=%q, want %q", resp.Version, req.WriteVersion)
	}
	got := strings.ToLower(strings.TrimSpace(resp.MD5))
	want := strings.ToLower(strings.TrimSpace(req.WriteMD5))
	if got != want {
		t.Fatalf("MD5=%q, want %q", resp.MD5, req.WriteMD5)
	}
	if resp.WrittenJS != "" {
		assertContainsAll(t, resp.WrittenJS,
			"BROWSER_AGENT_BUNDLE_VERSION",
			"BROWSER_AGENT_BUNDLE_MD5",
			req.WriteVersion,
		)
	}
	assertExitZero(t, resp)
}
```
