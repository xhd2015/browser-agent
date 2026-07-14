## Expected Output

Success **stdout** is **exactly** the session directory path and a trailing
newline (no banners). Stderr is free-form milestone lines (token-asserted below).

```
---
version: 2
__PATH__: type=string, example=/tmp/browser-trace-base/2026-07-11-12-00-00-abc, session directory path
---
__PATH__
```

## Expected

- Exit code 0.
- Stdout is exactly `{SessionDir}\n`.
- Stderr contains default info milestones: listen/addr, session, ready/waiting,
  recording (token match, case-insensitive).
- `{sessionDir}/browser-trace.log` exists and is non-empty.
- Log file content overlaps milestone tokens (at least listen/session/recording
  family).

## Side Effects

- `browser-trace.log` written under the session directory.
- `recording.har` / `meta.json` may exist (not the focus of this leaf).

## Errors

- None.

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
	assertDefaultInfoMilestones(t, resp.Stderr)

	if !resp.LogFileExists {
		t.Fatalf("expected browser-trace.log at %q; file missing", resp.LogFilePath)
	}
	if len(resp.LogFile) == 0 {
		t.Fatal("browser-trace.log exists but is empty")
	}
	// Log file should mirror info-level content (same token families).
	logLow := strings.ToLower(string(resp.LogFile))
	hasUseful := strings.Contains(logLow, "listen") ||
		strings.Contains(logLow, "session") ||
		strings.Contains(logLow, "ready") ||
		strings.Contains(logLow, "recording") ||
		strings.Contains(logLow, "browser-trace")
	if !hasUseful {
		t.Fatalf("browser-trace.log missing expected milestone tokens; content:\n%s", resp.LogFile)
	}
}
```
