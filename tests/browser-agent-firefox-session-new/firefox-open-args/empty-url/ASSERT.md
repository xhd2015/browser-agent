## Expected

- Nil error from Run.
- FirefoxArgs is `[-new-window]` only (or at least starts with `-new-window` and no URL).
- FirefoxArgs does **not** contain `--load-extension` or `--user-data-dir`.
- No `http://` / `https://` argument (empty URL → no navigation URL).

## Side Effects

- None (pure function).

## Errors

- Presence of managed flags or a URL arg fails.

## Exit Code

- N/A (package API).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if len(resp.FirefoxArgs) == 0 || resp.FirefoxArgs[0] != "-new-window" {
		t.Fatalf("empty-url should start with -new-window; args=%v", resp.FirefoxArgs)
	}
	assertNoManagedFirefoxFlags(t, resp.FirefoxArgs)
	for _, a := range resp.FirefoxArgs {
		if strings.HasPrefix(a, "http://") || strings.HasPrefix(a, "https://") {
			t.Fatalf("empty-url must not include navigation URL; found %q in %v", a, resp.FirefoxArgs)
		}
	}
}
```
