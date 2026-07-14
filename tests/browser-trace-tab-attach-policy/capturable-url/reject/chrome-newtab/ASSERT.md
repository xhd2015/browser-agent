## Expected

Requirement scenario **#2** — chrome new tab:

- `IsCapturableTabURL("chrome://newtab/")` returns **false**.
- `ShouldAttemptAttach(true, true, false, "chrome://newtab/")` returns **false**.

## Side Effects

- None (pure functions).

## Errors

- Run must not return an error.
- True would attempt attach on a chrome-internal page (fails / skipped).

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
