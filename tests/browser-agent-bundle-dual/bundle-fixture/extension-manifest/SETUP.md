# Scenario

**Feature**: Bundle UseFixture stages extension with manifest + version (A1)

```
Bundle(UseFixture) -> ExtensionDir/manifest.json
  version field non-empty
```

## Preconditions

- ModeBundle with temp BundleRoot and fixture sources (parent Setup).
- BundlePasses = 1.

## Steps

1. Default single Bundle pass (parent defaults BundlePasses=1).

## Context

- Assert focuses on ExtensionDir + manifest version only.

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
