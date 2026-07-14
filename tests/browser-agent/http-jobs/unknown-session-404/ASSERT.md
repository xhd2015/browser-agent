## Expected

Requirement **C3**:

- HTTP status **404**.
- Body should indicate not found / unknown session (soft: status is mandatory).

## Side Effects

- No job executed for the live session.

## Errors

- 200 with ok=false is insufficient — unknown session must be 404.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusNotFound)
}
```
