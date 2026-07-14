## Expected

Requirement scenario **#7** — control page attach OK:

- `IsCapturableTabURL("http://127.0.0.1:43759/go")` returns **true**.
- `ShouldAttemptAttach(true, true, false, same URL)` returns **true**.

## Side Effects

- None (pure functions).

## Errors

- Run must not return an error.
- False would skip attaching the session/control tab inconsistently with
  `attachAllTabsInWindow` (which only skips chrome/extension/devtools).

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
