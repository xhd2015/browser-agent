## Expected

- Each generated id matches `^sess-[a-z0-9]{6}$`.
- Two consecutive calls produce different ids (retry up to 5 times if collision).

## Side Effects

- None (pure).

## Errors

- Format mismatch or identical ids after retries fails this leaf.

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
	assertValidSessionIDFormat(t, resp.GeneratedID1)
	assertValidSessionIDFormat(t, resp.GeneratedID2)
	if resp.GeneratedID1 == resp.GeneratedID2 {
		t.Fatalf("expected different ids; both %q", resp.GeneratedID1)
	}
	assertExitZero(t, resp)
}
```