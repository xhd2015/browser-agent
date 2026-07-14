## Expected

Requirement scenario **#3** — chrome-extension page:

- `IsCapturableTabURL("chrome-extension://abcdefghijklmnopqrstuvwxyz123456/popup.html")`
  returns **false**.
- `ShouldAttemptAttach` with open gates on that URL returns **false**.

## Side Effects

- None (pure functions).

## Errors

- Run must not return an error.
- Attaching to extension pages is unsupported / unsafe for this product.

## Exit Code

- Not applicable (library call).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPolicy(t, req, resp, err)
}
```
