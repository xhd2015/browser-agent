## Expected

Requirement scenario **#4** — HTTPS app URL + open gates:

- `IsCapturableTabURL("https://app.example.com/app/weekly")` returns **true**.
- `ShouldAttemptAttach(true, true, false, same URL)` returns **true**.

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