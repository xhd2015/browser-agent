## Expected

- Exit 0; stdout contains sess-test and both tab HAR filenames

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.CLIErr != "" {
		t.Fatalf("CLIErr=%q stderr=%q", resp.CLIErr, resp.Stderr)
	}
	assertExitZero(t, resp)
	out := resp.Stdout
	for _, want := range []string{"sess-test", "tab-1-app.har", "tab-2-quiet.har"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, truncate(out, 800))
		}
	}
}
```
