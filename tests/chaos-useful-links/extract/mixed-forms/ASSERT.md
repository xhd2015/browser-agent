## Expected

- LoadSeedsFromFile succeeds (ResolveErr nil).
- Source.Type is `links`; Source.Path ends with `mixed.md`; SHA256 non-empty.
- At least 4 seeds; exact count 4 preferred (four distinct URLs).
- Seed URLs contain bare, md-link, backtick, and angle path segments.
- Each seed has non-empty ID and URL.

## Side Effects

- None (pure file read + parse).

## Errors

- Missing package/API or wrong extract rules fails this leaf.

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertResolveOK(t, resp)

	if resp.Resolved.Source.Type != "links" {
		t.Fatalf("Source.Type=%q want links", resp.Resolved.Source.Type)
	}
	if !strings.HasSuffix(resp.Resolved.Source.Path, "mixed.md") &&
		filepath.Base(resp.Resolved.Source.Path) != "mixed.md" {
		t.Fatalf("Source.Path=%q want basename mixed.md", resp.Resolved.Source.Path)
	}
	if strings.TrimSpace(resp.Resolved.Source.SHA256) == "" {
		t.Fatal("Source.SHA256 empty; want file content hash")
	}

	if req.WantCountSet && len(resp.Seeds) != req.WantCount {
		t.Fatalf("seed count=%d want %d; urls=%v", len(resp.Seeds), req.WantCount, seedURLs(resp.Seeds))
	}
	if len(resp.Seeds) < 4 {
		t.Fatalf("seed count=%d want ≥4; urls=%v", len(resp.Seeds), seedURLs(resp.Seeds))
	}
	assertWantURLs(t, resp.Seeds, req.WantURLs)

	for _, s := range resp.Seeds {
		if strings.TrimSpace(s.ID) == "" {
			t.Fatalf("seed missing ID: %+v", s)
		}
		if strings.TrimSpace(s.URL) == "" {
			t.Fatalf("seed missing URL: %+v", s)
		}
	}
}
```
