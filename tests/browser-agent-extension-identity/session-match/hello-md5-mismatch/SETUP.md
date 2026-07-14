# Scenario

**Feature**: hello md5 mismatch → md5_mismatch (C4)

```
serve (embedded md5 M)
Fake Extension hello {version: embedded, bundle_md5: fff…f}
GET /v1/session -> extension_match = md5_mismatch
```

## Preconditions

- ForceHelloMD5 is a 32-hex value distinct from embedded.

## Steps

1. Set SessionMatchKind = hello-md5-mismatch.
2. ForceHelloMD5 = ffffffffffffffffffffffffffffffff.

## Context

- Requirement C4.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionMatchKind = SessionMatchMD5Mismatch
	req.DoHello = true
	req.ForceHelloMD5 = "ffffffffffffffffffffffffffffffff"
	req.HelloFeatures = []string{"browser-agent"}
	return nil
}
```
