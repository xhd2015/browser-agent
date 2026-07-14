# Scenario

**Feature**: EnsureCanonicalExtension canonical layout (nested root)

```
TestHome -> DefaultExtensionInstallLayout()
Test Client -> EnsureCanonicalExtension() -> extensions/browser-agent/{version}/
```

## Preconditions

- Nested root; parent `Run` not inherited.
- Isolated `TestHome` per leaf via `t.Setenv("HOME", …)`.

## Steps

1. Allocate `TestHome` temp dir.
2. Leaf Setup sets `CanonicalPathOp`.

## Context

- **Compile RED** until `EnsureCanonicalExtension` / `DefaultExtensionInstallLayout` exist.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	dir := t.TempDir()
	req.TestHome = filepath.Join(dir, "home")
	return os.MkdirAll(req.TestHome, 0o755)
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertCanonicalPathSegment(t *testing.T, path string) {
	t.Helper()
	norm := filepath.ToSlash(path)
	if !strings.Contains(norm, "extensions/browser-agent/") {
		t.Fatalf("path should contain extensions/browser-agent/; got %q", path)
	}
}
```