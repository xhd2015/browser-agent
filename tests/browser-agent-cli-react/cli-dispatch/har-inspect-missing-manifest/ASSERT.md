## Expected

- Non-zero exit / CLIErr mentions manifest.json

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.CLIErr == "" {
		t.Fatal("expected CLIErr for missing manifest")
	}
	if !strings.Contains(resp.CLIErr, "manifest.json") {
		t.Fatalf("CLIErr=%q", resp.CLIErr)
	}
	if resp.ExitCode == 0 {
		t.Fatal("expected non-zero exit")
	}
}
```
