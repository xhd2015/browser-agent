# Scenario

**Feature**: invalid / unparseable request URLs bucket as host `"opaque"` (#7)

```
# Attached tab with mix of good and bad URLs
Background: AttachedTabIds={3}
  good: https://ok.example.com/a
  bad:  "not a url", "", "http://[:::"  (unparseable)
Test Client -> BuildPopupStats
          -> no panic / no error
          -> bad URLs contribute host "opaque"
          -> domainCount includes opaque once + ok.example.com
```

## Preconditions

- One attached tab id `3`.
- At least one valid URL and at least two invalid/unparseable URLs.
- Opaque bucket name is the literal `opaque` (root helper `OpaqueHost`).

## Steps

1. Attach tab `3` with title `"Mixed"`.
2. Add entries:
   - valid: `https://ok.example.com/a`, `https://ok.example.com/b`
   - invalid: `not a url`, empty string `""`, `http://[:bad`
3. Total entries = 5.

## Context

- Builder must never panic on bad URLs; all bad hosts collapse to one domain
  bucket so chips/domain pills stay stable.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	attachTabs(req, 3)
	setTabMeta(req, 3, "Mixed", "https://ok.example.com/", true)

	addEntry(req, 3, "good-1", "https://ok.example.com/a")
	addEntry(req, 3, "good-2", "https://ok.example.com/b")
	addEntry(req, 3, "bad-1", "not a url")
	addEntry(req, 3, "bad-2", "")
	addEntry(req, 3, "bad-3", "http://[:bad")
	return nil
}
```
