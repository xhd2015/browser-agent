## Expected

Requirement scenario **#5** — already attached:

- `IsCapturableTabURL(CapturableFixture)` returns **true**.
- `ShouldAttemptAttach(true, true, true, CapturableFixture)` returns **false**.

## Side Effects

- None (pure functions).

## Errors

- Run must not return an error.
- True would re-enter attach for a tab already in `attachedTabs`.

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
