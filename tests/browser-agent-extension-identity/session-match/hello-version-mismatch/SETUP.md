# Scenario

**Feature**: hello version mismatch → version_mismatch + warning (C3)

```
serve (embedded version V)
Fake Extension hello {version: 9.9.9, bundle_md5: embedded}
GET /v1/session -> extension_match = version_mismatch
stderr warning contains embedded + loaded version tokens
```

## Preconditions

- ForceHelloVersion differs from embedded (harness default 9.9.9).

## Steps

1. Set SessionMatchKind = hello-version-mismatch.
2. ForceHelloVersion = 9.9.9.

## Context

- Requirement C3. Jobs must not hard-fail solely due to mismatch (not asserted here).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionMatchKind = SessionMatchVersionMismatch
	req.DoHello = true
	req.ForceHelloVersion = "9.9.9"
	req.HelloFeatures = []string{"browser-agent"}
	return nil
}
```
