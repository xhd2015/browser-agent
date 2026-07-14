## Expected

Requirement scenario **#6** (not recording):

- `IsCapturableTabURL(CapturableFixture)` returns **true**.
- `ShouldAttemptAttach(false, true, false, CapturableFixture)` returns **false**.

## Side Effects

- None (pure functions).

## Errors

- Run must not return an error.
- True would attach while the agent is not recording (wrong lifecycle).

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
