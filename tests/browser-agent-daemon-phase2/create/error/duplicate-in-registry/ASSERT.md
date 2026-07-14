## Expected

- First `Create("dup-reg")` succeeds (Run pre-step).
- Second `Create("dup-reg")` returns error with `errors.Is(err, ErrSessionExists)`.

## Side Effects

- Single session dir remains from first Create.

## Errors

- Second create without ErrSessionExists fails this leaf.

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
	assertErrSessionExists(t, resp.SecondCreateErr)
	assertExitZero(t, resp)
}
```