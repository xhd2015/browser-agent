## Expected

After implementer lands wait-extension (**RED** on current code):

- Exit code non-zero.
- Error message contains "does not support browser-agent".
- Extension connected but features/version rejected.

## Side Effects

- Session created but wait rejected due to unsupported extension.
- Extension WS connection established.

## Errors

- Exit code 0 (no error for unsupported extension) fails.
- Error message missing or incorrect fails.

## Exit Code

- Non-zero (1).

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}

	assertExitNonZero(t, resp)

	assertContains(t, resp.ErrStr, "does not support browser-agent")
}
```
