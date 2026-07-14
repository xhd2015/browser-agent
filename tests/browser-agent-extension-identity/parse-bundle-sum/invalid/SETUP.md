# Scenario

**Feature**: parse missing/invalid bundle-sum.js → error (A2)

```
empty / garbage / missing VERSION+MD5 tokens
  -> ParseBundleSumJS -> non-nil error
```

## Preconditions

- BundleSumJS is not a valid generated sum file.

## Steps

1. Set BundleSumJS to invalid content (no version/md5 assignments).

## Context

- Requirement A2. One representative invalid payload is enough; empty and
  token-less garbage both must error under the same contract.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Deliberately missing BROWSER_AGENT_BUNDLE_* assignments.
	req.BundleSumJS = []byte("// not a bundle-sum\nconsole.log('nope');\n")
	return nil
}
```
