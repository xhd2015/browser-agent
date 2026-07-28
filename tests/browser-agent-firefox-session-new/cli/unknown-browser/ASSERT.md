## Expected

- Nonzero exit (`ExitCode != 0`) and/or non-empty `CLIErr`.
- Error text (CLIErr and/or Stderr) mentions `unknown browser` (case-insensitive)
  and includes the invalid value `safari`.

## Side Effects

- No session create required; fail-fast preferred.

## Errors

- Success (exit 0 / empty CLIErr) fails the leaf.

## Exit Code

- Nonzero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	// Run returns nil transport error for expected failure path.
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode == 0 && resp.CLIErr == "" {
		t.Fatalf("expected unknown browser to fail; ExitCode=0 CLIErr empty; stdout=%q stderr=%q",
			truncate(resp.Stdout, 200), truncate(resp.Stderr, 200))
	}
	combined := resp.CLIErr + "\n" + resp.Stderr + "\n" + resp.Stdout
	low := strings.ToLower(combined)
	if !strings.Contains(low, "unknown browser") {
		t.Fatalf("error should mention unknown browser; got:\n%s", truncate(combined, 800))
	}
	want := req.UnknownBrowser
	if want == "" {
		want = "safari"
	}
	if !strings.Contains(low, strings.ToLower(want)) {
		t.Fatalf("error should include browser value %q; got:\n%s", want, truncate(combined, 800))
	}
}
```
