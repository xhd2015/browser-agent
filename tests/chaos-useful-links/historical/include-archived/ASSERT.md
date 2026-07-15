## Expected

- All 7 fixture URLs present (3 active + 4 old).
- Counts.ArchivedSkipped == 0.

## Side Effects

- None.

## Errors

- Still skipping archived sections when IncludeArchived=true fails this leaf.

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

	if resp.Resolved.Counts.ArchivedSkipped != 0 {
		t.Fatalf("Counts.ArchivedSkipped=%d want 0 when IncludeArchived", resp.Resolved.Counts.ArchivedSkipped)
	}
}
```
