# Scenario

**Feature**: content md5 ignores bundle-sum.js (G2 exclude rule)

```
h1 = Compute(dir without sum)
WriteBundleSumJS(dir, version, h1)  # or raw write of sum file
h2 = Compute(dir with bundle-sum.js)
  -> h1 == h2
```

## Preconditions

- Default fixture has no `bundle-sum.js` initially.

## Steps

1. Set ComputeProbe = ignores-bundle-sum.

## Context

- Extension of B1/B2 covering the documented skip of `bundle-sum.js`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ComputeProbe = ComputeProbeIgnoresBundleSum
	req.WriteVersion = "1.0.1"
	return nil
}
```
