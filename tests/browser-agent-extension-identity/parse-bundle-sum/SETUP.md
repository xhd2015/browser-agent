# Scenario

**Feature**: ParseBundleSumJS — SW-loadable bundle-sum.js → BundleSum

```
# parse valid or invalid generated JS
Test Client -> ParseBundleSumJS(data)
  -> BundleSum{Version, MD5} | error
```

## Preconditions

- Mode = parse-bundle-sum.
- Leaves set BundleSumJS fixture bytes.

## Steps

1. Set Mode to parse-bundle-sum.
2. Leave BundleSumJS for leaf Setup.

## Context

- Requirement G1 / scenarios A1–A2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeParseBundleSum
	return nil
}
```
