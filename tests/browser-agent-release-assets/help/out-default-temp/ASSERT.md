## Expected Output

Help text describes `--out` as optional with a **temp** default, and must not claim
that `--out` is required for pack.

```text
version: 2
# --out … temp … (default); must not say required for pack
```

## Expected

- Exit code **0**.
- Combined stdout+stderr is non-empty and mentions **`--out`**.
- Combined text (case-insensitive) contains **`temp`** (temp dir / temporary default).
- Combined text does **not** contain **`required for pack`** (old wording).
- Soft ban: avoid presenting `--out` solely as a required flag in the Flags block
  with the exact legacy phrase above (hard-checked).
- Trailing newline on stdout or stderr (same rule as `help/mentions-upload`).

## Side Effects

- None (help only).

## Errors

- Non-zero exit, empty help, missing `--out`, missing temp wording, or legacy
  “required for pack” phrase fails.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertExitZero(t, resp)

	out := combinedOut(resp)
	if strings.TrimSpace(out) == "" {
		t.Fatalf("help output empty; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(out, "--out") {
		t.Fatalf("help missing --out; out=%s", truncate(out, 600))
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "temp") {
		t.Fatalf("help should say --out defaults to a temp dir; out=%s", truncate(out, 600))
	}
	if strings.Contains(lower, "required for pack") {
		t.Fatalf("help must not say --out is required for pack; out=%s", truncate(out, 600))
	}

	switch {
	case resp.Stdout != "" && strings.HasSuffix(resp.Stdout, "\n"):
		// ok
	case resp.Stderr != "" && strings.HasSuffix(resp.Stderr, "\n"):
		// ok
	default:
		t.Fatalf("help text must end with trailing newline on stdout or stderr; stdout=%q stderr=%q",
			truncate(resp.Stdout, 200), truncate(resp.Stderr, 200))
	}
}
```
