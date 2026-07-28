## Expected

- `background.js` found.
- Job type token **`create_tab`** present (quoted / case form preferred).
- Creates tabs via **`tabs.create`** (`browser.tabs.create` / `chrome.tabs.create` / `.tabs.create`).

## Side Effects

- None.

## Errors

- Missing create_tab branch or tabs.create fails create path contract.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !jobTypeTokenPresent(src, "create_tab") {
		t.Fatalf("background missing create_tab job branch token; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	if !hasTabsCreate(src) {
		t.Fatalf("create_tab must use tabs.create (browser/chrome); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
}
```
