# Scenario

**Feature**: Package Bundle with UseFixture under isolated temp Root

```
# stage extension + session-page fixtures into temp Root embed paths
Test Client -> Bundle(Root=temp, UseFixture=true, Fixture* abs from ModuleRoot)
  -> ExtensionDir + SessionPageDir under temp Root
```

## Preconditions

- ModeBundle.
- BundleRoot is a fresh temp directory (not ModuleRoot).
- UseFixture=true.
- Fixture sources resolved from ModuleRoot (absolute paths).
- BundlePasses defaults to 1; idempotent leaf sets 2.

## Steps

1. Set Mode = ModeBundle.
2. Allocate BundleRoot = t.TempDir().
3. Resolve FixtureExtensionDir and FixtureSessionPageDir from ModuleRoot.
4. Set UseFixture=true.
5. Leave BundlePasses / assertion focus to leaves.

## Context

- Must not mutate live `browseragent/embedded/**` under ModuleRoot.
- No npm; fixture copy only.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeBundle
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	req.BundleRoot = t.TempDir()
	if err := os.MkdirAll(req.BundleRoot, 0o755); err != nil {
		return err
	}
	req.UseFixture = true
	if req.FixtureExtensionDir == "" {
		req.FixtureExtensionDir = resolveFixtureExtensionDir(req.ModuleRoot)
	}
	if req.FixtureSessionPageDir == "" {
		req.FixtureSessionPageDir = resolveFixtureSessionPageDir(req.ModuleRoot)
	}
	// Sanity: fixtures must exist so RED is "Bundle missing", not "source missing".
	if _, err := os.Stat(filepath.Join(req.FixtureExtensionDir, "manifest.json")); err != nil {
		t.Fatalf("fixture extension missing manifest at %s: %v", req.FixtureExtensionDir, err)
	}
	foundSession := false
	for _, name := range []string{"index.html", "session-page.html"} {
		if _, err := os.Stat(filepath.Join(req.FixtureSessionPageDir, name)); err == nil {
			foundSession = true
			break
		}
	}
	if !foundSession {
		t.Fatalf("fixture session-page missing index under %s", req.FixtureSessionPageDir)
	}
	if req.BundlePasses <= 0 {
		req.BundlePasses = 1
	}
	return nil
}
```
