## Expected

Requirement **B2**:

- `MD5First` and `MD5Second` are both valid 32-hex digests.
- They are **not** equal after mutating one file.
- ExitCode 0.

## Side Effects

- Mutated file only inside leaf temp dir.

## Errors

- Equal digests after edit fail the leaf (hash not content-sensitive).

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
	if resp.MD5First == resp.MD5Second {
		t.Fatalf("md5 unchanged after file edit: %q", resp.MD5First)
	}
	if resp.MD5Equal {
		t.Fatal("MD5Equal=true after edit")
	}
	assertExitZero(t, resp)
}
```
