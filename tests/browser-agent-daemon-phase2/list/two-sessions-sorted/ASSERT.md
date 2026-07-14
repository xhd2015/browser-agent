## Expected

- After creating `sess-bbb` then `sess-aaa`, `List()` returns ids `["sess-aaa", "sess-bbb"]`.

## Side Effects

- Two session dirs created.

## Errors

- Wrong count or order fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	want := []string{"sess-aaa", "sess-bbb"}
	if len(resp.ListSessionIDs) != len(want) {
		t.Fatalf("List len=%d want %d got %v", len(resp.ListSessionIDs), len(want), resp.ListSessionIDs)
	}
	for i := range want {
		if resp.ListSessionIDs[i] != want[i] {
			t.Fatalf("List[%d]=%q want %q full=%v", i, resp.ListSessionIDs[i], want[i], resp.ListSessionIDs)
		}
	}
	assertExitZero(t, resp)
}
```