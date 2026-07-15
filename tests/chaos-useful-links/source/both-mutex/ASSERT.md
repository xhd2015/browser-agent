## Expected

- ResolveErr non-nil; ExitCode non-zero.
- Must not silently prefer one source.

## Side Effects

- None.

## Errors

- Accepting both sources fails this leaf.

## Exit Code

- Non-zero.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertResolveErr(t, resp)
}
```
