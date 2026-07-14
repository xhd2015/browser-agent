# Scenario

**Feature**: ComputeExtensionContentMD5 — deterministic dir hash excluding bundle-sum.js

```
# stage fixture extension dir
Test Client -> write manifest/background/content/popup
Test Client -> ComputeExtensionContentMD5(dir) -> 32-hex md5
  # skips basename bundle-sum.js
```

## Preconditions

- Mode = compute-content-md5.
- Leaves set ComputeProbe.

## Steps

1. Set Mode to compute-content-md5.
2. Leave ComputeProbe / edit fields for leaves.

## Context

- Requirement G2 / scenarios B1–B2 (+ exclude-sum).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeComputeContentMD5
	return nil
}
```
