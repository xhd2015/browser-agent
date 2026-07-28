## Expected

Requirement **W3** (dual install paths kept):

- Combined SessionPageApp + InstallGuideline sources still include:
  - Firefox: **`about:debugging`**
  - Chrome: **`chrome://extensions`**
- Soft: temporary add-on and/or Load unpacked still present.

## Side Effects

- None (read-only FS).

## Errors

- Removing either chrome or firefox install path fails.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	src := resp.CombinedText
	if strings.TrimSpace(src) == "" {
		src = assertSourcePresent(t, req, resp, "SessionPageApp+InstallGuideline")
	}

	if !strings.Contains(src, "about:debugging") {
		t.Fatalf("dual-path keep: must retain about:debugging (Firefox install); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	if !strings.Contains(src, "chrome://extensions") {
		t.Fatalf("dual-path keep: must retain chrome://extensions (Chrome install); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
}
```
