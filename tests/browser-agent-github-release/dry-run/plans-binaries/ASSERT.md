## Expected

- Exit code **0** (dry-run soft-warns dirty git / missing creds).
- Stdout contains **`[dry-run]`** and **`would build:`**.
- At least one planned binary matching `browser-agent-v` + `darwin` or `linux`.
- Mentions hydrate pack unless skipped (optional soft: `would pack hydrate` or `session-page`).

```go
import (
	"regexp"
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
	if !strings.Contains(out, "[dry-run]") {
		t.Fatalf("missing [dry-run] prefix; out=%s", truncate(out, 700))
	}
	if !strings.Contains(out, "would build:") {
		t.Fatalf("missing would build:; out=%s", truncate(out, 700))
	}
	// browser-agent-v0.3.1-darwin-arm64 style
	re := regexp.MustCompile(`browser-agent-v[0-9][0-9A-Za-z._+-]*-(darwin|linux)-(amd64|arm64)`)
	if !re.MatchString(out) {
		t.Fatalf("no planned binary name browser-agent-v*-{os}-{arch}; out=%s", truncate(out, 900))
	}
	// Prefer all four specs present
	for _, pair := range []string{"darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64"} {
		if !strings.Contains(out, pair) {
			t.Fatalf("dry-run missing platform %s; out=%s", pair, truncate(out, 900))
		}
	}
}
```
