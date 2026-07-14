# Scenario

**Feature**: multi-tab merged HAR + meta save, session dir naming, stdout trailing newline

```
# Multi-tab capture contract
Mock Extension complete.har.entries from tab 1 and tab 2 (_tabId)
Control Server merges/persists entries (sorted by startedDateTime)

# Storage layout
BaseDir / YYYY-MM-DD-HH-MM-SS-<suffix> / meta.json
BaseDir / YYYY-MM-DD-HH-MM-SS-<suffix> / recording.har

# User-facing stdout ends with \n after last content line (session path)
browser-trace stdout -> "...<session-dir>\n"
```

## Preconditions

- Extension-stop happy path with default multi-tab mock HAR (≥2 entries, different `_tabId`).
- Optional fixture `testdata/multi-tab.har` may be loaded when present.

## Steps

1. If `testdata/multi-tab.har` exists beside this leaf, load it into `req.MockHAR`.
2. Otherwise leave `MockHAR` empty so `Run` uses built-in multi-tab sample.
3. Keep extension stop mode and record-and-complete script.

## Context

- Requirement scenarios #4 (multi-tab HAR + meta), #7 (dir naming), #8 (stdout `\n`).

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtRecordAndComplete
	req.StopMode = StopExtension
	req.MockStopReason = "extension"
	if b, err := os.ReadFile("testdata/multi-tab.har"); err == nil && len(b) > 0 {
		req.MockHAR = b
	}
	return nil
}
```
