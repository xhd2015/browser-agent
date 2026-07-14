## Expected

Requirement scenario **#1** — empty URL:

- `IsCapturableTabURL("")` returns **false**.
- `ShouldAttemptAttach(true, true, false, "")` returns **false**.

## Side Effects

- None (pure functions).

## Errors

- Run must not return an error.
- Returning true would attempt attach on a blank new-tab shell (fails / noise).

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
