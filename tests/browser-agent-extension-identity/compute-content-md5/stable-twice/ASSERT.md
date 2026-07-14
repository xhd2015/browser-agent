## Expected

Requirement **B1**:

- First and second `ComputeExtensionContentMD5` results are equal.
- Each is 32 lowercase hex characters.
- ExitCode 0; no ErrText.

## Side Effects

- Fixture dir under temp only.

## Errors

- Hash inequality or non-hex digest fails the leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ErrText != "" {
		t.Fatalf("Compute error: %s", resp.ErrText)
	}
	assertHexMD5(t, resp.MD5First, "MD5First")
	assertHexMD5(t, resp.MD5Second, "MD5Second")
	if !resp.MD5Equal || resp.MD5First != resp.MD5Second {
		t.Fatalf("md5 not stable: first=%q second=%q", resp.MD5First, resp.MD5Second)
	}
	assertExitZero(t, resp)
}
```
