# Scenario

**Feature**: session-page HTML without assets JS is incomplete

```
testdata/session-page-html-only (index.html only, no assets/)
  -> EmbedCompleteFS(fs, "session-page") -> false
```

## Preconditions

- Fixture has non-empty index.html but no `assets/*.js` and no script src
  that exists in the FS.

## Steps

1. Set `FixtureName = FixtureSessionPageHTMLOnly`.
2. Set `ExpectComplete = false`.

## Context

- Incomplete embed shape (e.g. partial go install staging).
- Leaf dir is `html-only-no-assets` (not `*-js`) so generated
  `*_test.go` is not constrained to GOOS=js.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FixtureName = FixtureSessionPageHTMLOnly
	req.ExpectComplete = false
	return nil
}
```
