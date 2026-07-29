## Expected

- Exit code **0**.
- Help mentions **`--dry-run`**.
- Help mentions **`install.sh`** or raw.githubusercontent / curl install path.
- Help mentions **`.upload-credentials`** or credentials.
- Trailing newline on stdout or stderr.

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
	assertExitZero(t, resp)
	out := combinedOut(resp)
	if strings.TrimSpace(out) == "" {
		t.Fatal("help empty")
	}
	if !strings.Contains(out, "--dry-run") {
		t.Fatalf("help missing --dry-run; out=%s", truncate(out, 700))
	}
	if !strings.Contains(out, "install.sh") && !strings.Contains(out, "raw.githubusercontent.com") {
		t.Fatalf("help missing install.sh / curl install; out=%s", truncate(out, 700))
	}
	if !strings.Contains(out, "upload-credentials") && !strings.Contains(out, "credentials") {
		t.Fatalf("help missing credentials note; out=%s", truncate(out, 700))
	}
	switch {
	case resp.Stdout != "" && strings.HasSuffix(resp.Stdout, "\n"):
	case resp.Stderr != "" && strings.HasSuffix(resp.Stderr, "\n"):
	default:
		t.Fatalf("help must end with newline; stdout=%q stderr=%q",
			truncate(resp.Stdout, 120), truncate(resp.Stderr, 120))
	}
}
```
