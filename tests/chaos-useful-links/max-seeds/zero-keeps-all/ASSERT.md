## Expected

- Exactly 5 seeds; all a–e path segments present.
- Counts.AfterFilter == 5; Counts.Deduped == 5.

## Side Effects

- None.

## Errors

- Capping when MaxSeeds=0 fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertResolveOK(t, resp)

	if len(resp.Seeds) != 5 {
		t.Fatalf("seed count=%d want 5; urls=%v", len(resp.Seeds), seedURLs(resp.Seeds))
	}
	assertWantURLs(t, resp.Seeds, req.WantURLs)

	if resp.Resolved.Counts.Deduped != 5 {
		t.Fatalf("Counts.Deduped=%d want 5", resp.Resolved.Counts.Deduped)
	}
	if resp.Resolved.Counts.AfterFilter != 5 {
		t.Fatalf("Counts.AfterFilter=%d want 5", resp.Resolved.Counts.AfterFilter)
	}
}
```
