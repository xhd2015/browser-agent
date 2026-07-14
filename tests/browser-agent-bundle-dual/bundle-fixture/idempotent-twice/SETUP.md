# Scenario

**Feature**: Bundle twice is idempotent with stable paths (A3)

```
Bundle(UseFixture) -> paths1
Bundle(UseFixture) -> paths2
  paths1 == paths2; both succeed
```

## Preconditions

- ModeBundle parent Setup complete.
- BundlePasses = 2.

## Steps

1. Set BundlePasses = 2 so Run invokes Bundle twice with same opts.

## Context

- Idempotent: second stage must not fail; destinations stable.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	if req.Mode != ModeBundle {
		t.Fatalf("Mode = %q, want %q", req.Mode, ModeBundle)
	}
	req.BundlePasses = 2
	return nil
}
```
