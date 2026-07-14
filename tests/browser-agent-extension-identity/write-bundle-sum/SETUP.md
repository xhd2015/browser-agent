# Scenario

**Feature**: WriteBundleSumJS — write generated bundle-sum.js under extension dir

```
Test Client -> WriteBundleSumJS(dir, version, md5)
  -> dir/bundle-sum.js
  -> ParseBundleSumJS(bytes) round-trips version+md5
```

## Preconditions

- Mode = write-bundle-sum.

## Steps

1. Set Mode to write-bundle-sum.
2. Leaves set WriteVersion / WriteMD5.

## Context

- Requirement G2 / scenario B3.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeWriteBundleSum
	return nil
}
```
