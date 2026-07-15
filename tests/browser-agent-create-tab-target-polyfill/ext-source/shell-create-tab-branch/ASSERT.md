## Expected

Requirement **E1**:

- background.js found under `Chrome-Ext-Browser-Agent`.
- Source contains job type token **`create_tab`** (quoted / case form preferred).
- Source contains **`chrome.tabs.create`** (shared create path with Target.createTarget).

## Side Effects

- None (read-only FS).

## Errors

- Missing create_tab branch or tabs.create fails shared create path contract.

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
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("shell background missing; err=%q found=%v ModuleRoot=%s",
			resp.ErrText, resp.FoundPaths, req.ModuleRoot)
	}
	src := resp.CombinedText
	if !jobTypeTokenPresent(src, "create_tab") {
		t.Fatalf("shell background missing create_tab job branch token; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	if !strings.Contains(src, "chrome.tabs.create") {
		t.Fatalf("shell background must use chrome.tabs.create for create path; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
}
```
