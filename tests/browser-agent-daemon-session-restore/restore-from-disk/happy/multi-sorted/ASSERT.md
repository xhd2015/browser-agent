## Expected

- Restore err nil.
- List session ids exactly `["sess-aaa", "sess-bbb"]` (sorted ascending).
- Both ids Get ok; both waiting_extension.

## Side Effects

- Both session dirs remain on disk.

## Errors

- Wrong order or missing id fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRestoreOK(t, resp)
	want := []string{"sess-aaa", "sess-bbb"}
	if len(resp.ListIDs) != len(want) {
		t.Fatalf("ListIDs=%v want %v", resp.ListIDs, want)
	}
	for i := range want {
		if resp.ListIDs[i] != want[i] {
			t.Fatalf("ListIDs=%v want %v", resp.ListIDs, want)
		}
	}
	assertGetOK(t, resp, "sess-aaa", true)
	assertGetOK(t, resp, "sess-bbb", true)
	assertWaitingExtension(t, resp, "sess-aaa")
	assertWaitingExtension(t, resp, "sess-bbb")
	assertExitZero(t, resp)
}
```
