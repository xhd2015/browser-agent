# Scenario

**Feature**: hello without bundle_md5 → md5_unknown (C5)

```
serve
Fake Extension hello {version: embedded}  # no bundle_md5 field
GET /v1/session -> extension_match = md5_unknown
```

## Preconditions

- HelloOmitMD5 true.

## Steps

1. Set SessionMatchKind = hello-md5-unknown.
2. HelloOmitMD5 = true.

## Context

- Requirement C5. Older extensions may omit md5.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionMatchKind = SessionMatchMD5Unknown
	req.DoHello = true
	req.HelloOmitMD5 = true
	req.HelloFeatures = []string{"browser-agent"}
	return nil
}
```
