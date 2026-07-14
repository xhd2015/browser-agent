# Scenario

**Feature**: parse valid bundle-sum.js → version + md5 (A1)

```
// browser-agent bundle-sum — generated; do not edit
var BROWSER_AGENT_BUNDLE_VERSION = "1.0.1";
var BROWSER_AGENT_BUNDLE_MD5 = "a1b2c3d4e5f6789012345678abcdef01";
  -> ParseBundleSumJS -> Version=1.0.1 MD5=a1b2…01
```

## Preconditions

- Fixture bytes match the documented generated format.

## Steps

1. Set BundleSumJS to valid fixture with known version + 32-hex md5.

## Context

- Requirement A1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.BundleSumJS = validBundleSumFixture("1.0.1", "a1b2c3d4e5f6789012345678abcdef01")
	return nil
}
```
