## Expected

- ResolveErr non-nil; ExitCode non-zero.
- No successful seed list required (Seeds should be empty/nil).

## Side Effects

- None.

## Errors

- Returning success without a source fails this leaf.

## Exit Code

- Non-zero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertResolveErr(t, resp)

	msg := strings.ToLower(resp.ResolveErrText)
	if !strings.Contains(msg, "links") && !strings.Contains(msg, "random") {
		t.Fatalf("error should mention links or random; got %q", resp.ResolveErrText)
	}
}
```
