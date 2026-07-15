## Expected

- `ResolveSessionPageIndexFS` succeeds (`err == nil`).
- `source` equals `"embed"`.
- `html` is non-empty and includes session root marker
  (`data-browser-agent-root` or `id="root"`).

## Side Effects

- None (read-only).

## Errors

- Non-nil resolve error, wrong source, or HTML without root marker fails.

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
	if resp.ResolveErr != nil {
		t.Fatalf("ResolveSessionPageIndexFS err=%v want nil; FSRoot=%s",
			resp.ResolveErr, resp.FSRoot)
	}
	if resp.Source != ResolveSourceEmbed {
		t.Fatalf("source=%q want %q", resp.Source, ResolveSourceEmbed)
	}
	assertHTMLHasSessionRoot(t, resp.HTML)
	// complete fixture index is real HTML, not a bare empty string
	if !strings.Contains(strings.ToLower(resp.HTML), "html") {
		t.Fatalf("HTML does not look like a document; body=%s",
			truncate(resp.HTML, 300))
	}
	assertExitZero(t, resp)
}
```
