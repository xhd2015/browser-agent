# Scenario

**Feature**: complete session-page fixture is reported complete

```
testdata/session-page-complete (index.html + assets/session-page.js)
  -> EmbedCompleteFS(fs, "session-page") -> true
```

## Preconditions

- Fixture `session-page-complete` has non-empty index.html and
  assets/session-page.js.

## Steps

1. Set `FixtureName = FixtureSessionPageComplete`.
2. Set `ExpectComplete = true`.

## Context

- Happy path for session-page completeness.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FixtureName = FixtureSessionPageComplete
	req.ExpectComplete = true
	return nil
}
```
