## Expected Output

Success-like stdout: session path and trailing newline.

```
---
version: 3
__PATH__: type=string, example=/tmp/browser-trace-base/2026-07-11-12-00-00-hb-session, session directory path
---
__PATH__
```

## Expected

Requirement scenario **#4** — heartbeat_lost with empty snapshot:

- Exit code **0**.
- Stdout includes session path and ends with `\n`.
- Stderr warning tokens: `warning`, `heartbeat`, and `unusual` **or** `acceptable`.
- `recording.har` exists as valid HAR with **0** entries.
- `meta.json` has stop_reason containing `heartbeat` and `partial: true`.
- Mock reached recording; did **not** POST entries; did **not** complete.

## Side Effects

- Partial session artifacts under session dir even with zero captured URLs.

## Errors

- Exit ≠ 0 is a failure.
- Non-empty HAR entries when never posted is a failure.
- Missing warning is a failure.

## Exit Code

- 0.

```go
import (
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
	assertHeartbeatWarning(t, resp.Stderr)

	stdout := resp.Stdout
	if lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n"); len(lines) == 1 {
		assert.Output(t, stdout, `---
version: 3
__PATH__: type=string, example=/tmp/browser-trace-base/2026-07-11-12-00-00-hb-session, session directory path
---
__PATH__
`)
	}

	assertHAREmptyOrMinimal(t, resp.HARJSON)
	assertMetaHeartbeatLost(t, resp.MetaJSON)

	if !resp.StatusRecording {
		t.Fatal("expected mock to reach recording status before silence")
	}
	if resp.EntriesPosted {
		t.Fatal("empty-snapshot leaf must not POST /v1/entries")
	}
	if resp.CompletePosted {
		t.Fatal("heartbeat_lost path should not POST /v1/complete")
	}
}
```
