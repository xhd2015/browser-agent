## Expected

- Exactly 2 seeds after dedupe.
- Both /same and /other present.
- Counts.Deduped == 2; Counts.Raw ≥ 3 (three same + one other before dedupe).

## Side Effects

- None.

## Errors

- Emitting duplicate /same seeds fails this leaf.

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
	assertWantURLs(t, resp.Seeds, req.WantURLs)

	sameN := 0
	for _, s := range resp.Seeds {
		if strings.Contains(s.URL, "example.com/same") {
			sameN++
		}
	}
	if sameN != 1 {
		t.Fatalf("same URL seed count=%d want 1; urls=%v", sameN, seedURLs(resp.Seeds))
	}

	c := resp.Resolved.Counts
	if c.Deduped != 2 {
		t.Fatalf("Counts.Deduped=%d want 2", c.Deduped)
	}
	if c.Raw < 3 {
		t.Fatalf("Counts.Raw=%d want ≥3 (duplicates before dedupe)", c.Raw)
	}
}
```
