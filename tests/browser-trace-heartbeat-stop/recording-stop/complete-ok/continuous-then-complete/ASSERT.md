## Expected Output

Success stdout: session path and trailing newline.

```
---
version: 2
__PATH__: type=string, example=/tmp/browser-trace-base/2026-07-11-12-00-00-hb-session, session directory path
---
__PATH__
```

## Expected

Requirement scenarios **#2** and **#5** — normal complete still works:

- Exit code **0**.
- Stdout includes session path and ends with `\n`.
- `recording.har` and `meta.json` exist.
- HAR contains the live URL from continuous entries/complete.
- Stderr must **not** look like a heartbeat_lost warning (no `heartbeat_lost`;
  no warning+heartbeat+unusual/acceptable combo).
- Mock posted complete; recording was reached.

## Side Effects

- Session artifacts under BaseDir; stop_reason is extension (or non-heartbeat) when present.
- `partial` should not be true for a normal complete (if field present).

## Errors

- Heartbeat_lost warning on this path is a failure.
- Missing complete / non-zero exit is a failure.

## Exit Code

- 0.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitZero(t, resp)
	assertSessionArtifactsExist(t, resp)
	assertStdoutSessionPath(t, resp)
	assertNoHeartbeatWarningRequired(t, resp.Stderr)

	stdout := resp.Stdout
	if lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n"); len(lines) == 1 {
		assert.Output(t, stdout, `---
version: 2
__PATH__: type=string, example=/tmp/browser-trace-base/2026-07-11-12-00-00-hb-session, session directory path
---
__PATH__
`)
	}

	urls := req.SnapshotURLs
	if len(urls) == 0 {
		urls = []string{"https://api.example.com/live"}
	}
	assertHARContainsURLs(t, resp.HARJSON, urls...)

	if !resp.StatusRecording {
		t.Fatal("expected recording status before complete")
	}
	if !resp.CompletePosted {
		t.Fatal("expected mock to POST /v1/complete")
	}

	if len(resp.MetaJSON) > 0 {
		var meta map[string]any
		if err := json.Unmarshal(resp.MetaJSON, &meta); err != nil {
			t.Fatalf("parse meta.json: %v", err)
		}
		if sr, ok := meta["stop_reason"].(string); ok && sr != "" {
			if strings.Contains(strings.ToLower(sr), "heartbeat") {
				t.Fatalf("meta.stop_reason = %q, want non-heartbeat (e.g. extension)", sr)
			}
		}
		if p, ok := meta["partial"].(bool); ok && p {
			t.Fatalf("meta.partial = true on normal complete path; want false or absent")
		}
	}
}
```
