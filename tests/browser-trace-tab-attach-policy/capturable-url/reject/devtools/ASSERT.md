## Expected

Skip-list alignment — devtools:

- `IsCapturableTabURL("devtools://devtools/bundled/inspector.html")` returns **false**.
- `ShouldAttemptAttach` with open gates on that URL returns **false**.

## Side Effects

- None (pure functions).

## Errors

- Run must not return an error.

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
