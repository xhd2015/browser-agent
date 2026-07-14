## Expected

Requirement **B5** (error variant):

- `CompleteErr` is non-empty.
- Message ideally mentions unknown / not found / no such (soft: any non-empty error OK).

## Side Effects

- None.

## Errors

- Silent success on unknown id is a failure for this MVP choice.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.CompleteErr == "" {
		t.Fatal("Complete unknown id must return an error (CompleteErr empty)")
	}
}
```
