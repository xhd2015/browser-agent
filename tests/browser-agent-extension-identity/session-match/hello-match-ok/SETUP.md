# Scenario

**Feature**: hello version+md5 match embedded → extension_match=ok (C2)

```
serve
Fake Extension hello {version=embedded, bundle_md5=embedded, features:[browser-agent]}
GET /v1/session
  -> extension_match = ok
  -> extension.connected = true
```

## Preconditions

- Harness reads embedded identity from first session probe, then hellos with same.

## Steps

1. Set SessionMatchKind = hello-match-ok.
2. DoHello true; features browser-agent.

## Context

- Requirement C2. Match=ok should not require orange mismatch warning.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionMatchKind = SessionMatchHelloOK
	req.DoHello = true
	req.HelloFeatures = []string{"browser-agent"}
	return nil
}
```
