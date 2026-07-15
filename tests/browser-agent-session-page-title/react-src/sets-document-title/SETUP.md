# Scenario

**Feature**: SessionPageApp sets `document.title` to `{sid} - Browser Agent` (T3)

```
react/src/ui/SessionPageApp.tsx (and/or session-page app)
  when sid known: document.title = sid + " - Browser Agent"
  when sid empty: do not set broken " - Browser Agent"
```

## Preconditions

- Mode already react-src from parent.

## Steps

1. Set `ReactProbe = ReactProbeDocumentTitle` (`sets-document-title`).

## Context

- Requirement scenarios 3–4 (known sid + missing/empty sid guard).
- Prefer `useEffect` on `sid`; string form may be template or concat.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeDocumentTitle
	return nil
}
```
