## Expected Output

Stdout exact path + newline; stderr should not carry progress milestones.

```
---
version: 3
__PATH__: type=string, example=/tmp/browser-trace-base/2026-07-11-12-00-00-abc, session directory path
---
__PATH__
```

## Expected

- Exit code 0.
- Stdout is exactly `{SessionDir}\n`.
- Stderr does **not** contain default info milestone phrases (listen on / ready
  wait / recording started). Benign empty stderr is ideal; non-fatal warnings
  without those tokens are acceptable.
- `browser-trace.log` is absent **or** empty (Quiet suppresses info+ file mirror).

## Side Effects

- No progress log pollution under Quiet.

## Errors

- None (success path).

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitZero(t, resp)
	assertStdoutSessionPathOnly(t, resp)

	// Quiet: must not emit info-style progress lines.
	// Allow empty stderr or warnings that do not look like lifecycle milestones.
	low := strings.ToLower(resp.Stderr)
	banned := []string{
		"listening",
		"listen on",
		"ready wait",
		"waiting for extension",
		"waiting for ready",
		"recording started",
		"start recording",
		"session url",
		"heartbeat",
	}
	for _, b := range banned {
		if strings.Contains(low, b) {
			t.Fatalf("Quiet stderr must not contain info token %q; stderr:\n%s", b, resp.Stderr)
		}
	}

	// Log file: Quiet should not leave info+ content.
	if resp.LogFileExists && len(resp.LogFile) > 0 {
		logLow := strings.ToLower(string(resp.LogFile))
		for _, b := range banned {
			if strings.Contains(logLow, b) {
				t.Fatalf("Quiet browser-trace.log must not contain info token %q; log:\n%s", b, resp.LogFile)
			}
		}
	}
}
```
