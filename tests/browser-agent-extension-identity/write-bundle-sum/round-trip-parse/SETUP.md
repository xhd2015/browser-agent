# Scenario

**Feature**: WriteBundleSumJS then Parse → round-trip (B3)

```
WriteBundleSumJS(dir, "1.0.1", "a1b2c3d4e5f6789012345678abcdef01")
  -> ParseBundleSumJS(read file)
  -> Version + MD5 match inputs
```

## Preconditions

- Harness stages default fixture dir then writes sum.

## Steps

1. Set WriteVersion and WriteMD5 to known values.

## Context

- Requirement B3.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WriteVersion = "1.0.1"
	req.WriteMD5 = "a1b2c3d4e5f6789012345678abcdef01"
	return nil
}
```
