## Expected

Requirement **A1**:

- Embed index HTML is non-empty.
- Root mount present: `id="root"` or `data-browser-agent-root` (or
  `browser-agent-root`).
- Embed path is non-empty (which file was read).

## Side Effects

- None (read-only embed FS).

## Errors

- Missing embed / empty HTML / no root mount fails.

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
	html := resp.EmbedHTML
	if strings.TrimSpace(html) == "" {
		html = resp.BodyString
	}
	if strings.TrimSpace(html) == "" {
		t.Fatalf("embedded session-page index HTML empty; path=%q listing=%v err=%q",
			resp.EmbedPath, resp.EmbedListing, resp.ErrText)
	}
	if !hasRootMount(html) {
		t.Fatalf("embed HTML missing root mount (id=root / data-browser-agent-root); path=%q body=%s",
			resp.EmbedPath, truncate(html, 600))
	}
	if resp.EmbedPath == "" {
		t.Fatal("EmbedPath empty (which index file was read?)")
	}
}
```
