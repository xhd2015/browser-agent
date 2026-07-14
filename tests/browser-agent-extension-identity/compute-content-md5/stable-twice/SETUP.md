# Scenario

**Feature**: ComputeMD5 same dir twice → equal (B1)

```
fixture dir
  -> ComputeExtensionContentMD5 x2
  -> h1 == h2; both 32-hex
```

## Preconditions

- Default fixture files staged by harness.

## Steps

1. Set ComputeProbe = stable-twice.

## Context

- Requirement B1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ComputeProbe = ComputeProbeStableTwice
	return nil
}
```
