## Expected Output

Success-like stdout: session path and trailing newline.

```
---
version: 2
__PATH__: type=string, example=/tmp/browser-trace-base/2026-07-11-12-00-00-hb-session, session directory path
---
__PATH__
```

## Expected

Requirement scenario **#3** — heartbeat_lost with entries snapshot:

- Exit code **0**.
- Stdout includes session path and ends with `\n`.
- Stderr warning tokens: `warning`, `heartbeat`, and `unusual` **or** `acceptable`.
- `recording.har` exists and contains both snapshot URLs.
- `meta.json` has stop_reason containing `heartbeat` (e.g. `heartbeat_lost`) and `partial: true`.
- Mock reached recording and posted entries; did **not** require complete.

## Side Effects

- Artifacts only under session dir: `recording.har`, `meta.json`.
- No requirement to write a non-zero exit or fail the process.

## Errors

- Exit ≠ 0 on heartbeat_lost is a failure (product: unusual but acceptable stop).
- HAR missing snapshot URLs is a failure.
- Missing partial meta is a failure.

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

	// Prefer strict path-only stdout when product prints only the path line.
	// Allow multi-line if implementer adds banners — fall back already covered
	// by assertStdoutSessionPath; try strict template when single-line path.
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
		urls = []string{
			"https://api.example.com/v1/alpha",
			"https://cdn.example.com/app.js",
		}
	}
	assertHARContainsURLs(t, resp.HARJSON, urls...)
	assertMetaHeartbeatLost(t, resp.MetaJSON)

	if !resp.StatusRecording {
		t.Fatal("expected mock to reach recording status before silence")
	}
	if !resp.EntriesPosted {
		t.Fatal("expected mock to POST /v1/entries before silence")
	}
	if resp.CompletePosted {
		t.Fatal("heartbeat_lost path should not require /v1/complete from mock")
	}

	// entry_count in meta should be >= number of snapshot URLs when present.
	// Soft: already covered by HAR URLs + partial.
}
```
