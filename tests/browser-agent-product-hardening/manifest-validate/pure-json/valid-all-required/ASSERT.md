## Expected

Requirement **A1**:

- `ValidateExtensionManifestJSON` returns nil (`ValidateOK` true).
- `ValidateErr` empty.
- ExitCode 0.

## Side Effects

- None (pure).

## Errors

- Any non-nil validate error fails this leaf.

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
	if !resp.ValidateOK {
		t.Fatalf("expected ValidateOK; ValidateErr=%q", resp.ValidateErr)
	}
	if resp.ValidateErr != "" {
		t.Fatalf("ValidateErr should be empty; got %q", resp.ValidateErr)
	}
	assertExitZero(t, resp)
}
```
