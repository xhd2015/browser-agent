# Scenario

**Feature**: Staged extension hosts control port 43761 (A4)

```
Bundle(UseFixture) -> ExtensionDir files
  content mentions 43761 (manifest hosts / matches)
```

## Preconditions

- ModeBundle parent Setup complete.
- BundlePasses = 1.

## Steps

1. Single Bundle; assert staged extension tree text contains 43761.

## Context

- Port must not be browser-trace 43759 for agent fixture.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	if req.Mode != ModeBundle {
		t.Fatalf("Mode = %q, want %q", req.Mode, ModeBundle)
	}
	req.BundlePasses = 1
	return nil
}
```
