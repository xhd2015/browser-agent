## Expected Output

Stdout remains exact session path + newline.

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
- Stderr still contains default info milestones.
- `{sessionDir}/browser-trace.log` does **not** exist.

## Side Effects

- No `browser-trace.log` under the session directory.
- Session dir may still contain HAR/meta.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitZero(t, resp)
	assertStdoutSessionPathOnly(t, resp)
	assertDefaultInfoMilestones(t, resp.Stderr)

	if resp.LogFileExists {
		t.Fatalf("NoLogFile=true but browser-trace.log exists at %q (%d bytes)",
			resp.LogFilePath, len(resp.LogFile))
	}
	if resp.LogFilePath != "" {
		if _, err := os.Stat(resp.LogFilePath); err == nil {
			t.Fatalf("NoLogFile=true but stat found log file %q", resp.LogFilePath)
		}
	}
}
```
