## Expected

- Active hosts present; old.* under historical/archived/deprecated absent.
- Exactly 3 active seeds.
- Counts.ArchivedSkipped ≥ 1.

## Side Effects

- None.

## Errors

- Including archived URLs by default fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertResolveOK(t, resp)

	if len(resp.Seeds) != req.WantCount {
		t.Fatalf("seed count=%d want %d; urls=%v", len(resp.Seeds), req.WantCount, seedURLs(resp.Seeds))
	}
	assertWantURLs(t, resp.Seeds, req.WantURLs)
	assertNotURLs(t, resp.Seeds, req.WantNotURLs)

	if resp.Resolved.Counts.ArchivedSkipped < 1 {
		t.Fatalf("Counts.ArchivedSkipped=%d want ≥1", resp.Resolved.Counts.ArchivedSkipped)
	}
}
```
