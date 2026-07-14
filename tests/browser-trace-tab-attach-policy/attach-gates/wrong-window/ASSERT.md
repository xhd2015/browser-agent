## Expected

Requirement scenario **#6** (wrong window):

- `IsCapturableTabURL(CapturableFixture)` returns **true**.
- `ShouldAttemptAttach(true, false, false, CapturableFixture)` returns **false**.

## Side Effects

- None (pure functions).

## Errors

- Run must not return an error.
- True would attach tabs outside the pinned capture window.

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
