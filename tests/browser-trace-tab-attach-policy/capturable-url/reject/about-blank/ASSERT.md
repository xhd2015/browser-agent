## Expected

Product prefer — about:blank:

- `IsCapturableTabURL("about:blank")` returns **false**.
- `ShouldAttemptAttach(true, true, false, "about:blank")` returns **false**.

## Side Effects

- None (pure functions).

## Errors

- Run must not return an error.
- True would thrash attach on intermediate blank documents.

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
