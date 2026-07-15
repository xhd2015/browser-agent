## Expected

- Exactly 2 seeds.
- Counts.Deduped == 5 (before cap); Counts.AfterFilter == 2.
- Each seed URL is one of the fixture seed.example.com/* URLs.

## Side Effects

- None.

## Errors

- Returning all 5 when MaxSeeds=2 fails this leaf.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertResolveOK(t, resp)

	if len(resp.Seeds) != 2 {
		t.Fatalf("seed count=%d want 2; urls=%v", len(resp.Seeds), seedURLs(resp.Seeds))
	}
	if resp.Resolved.Counts.Deduped != 5 {
		t.Fatalf("Counts.Deduped=%d want 5 (cap is after dedupe)", resp.Resolved.Counts.Deduped)
	}
	if resp.Resolved.Counts.AfterFilter != 2 {
		t.Fatalf("Counts.AfterFilter=%d want 2", resp.Resolved.Counts.AfterFilter)
	}
	for _, s := range resp.Seeds {
		if !strings.Contains(s.URL, "seed.example.com/") {
			t.Fatalf("unexpected seed URL %q", s.URL)
		}
	}
}
```
