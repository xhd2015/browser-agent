# Scenario

**Feature**: Bundle / StageFirefox writes `browseragent/embedded/extension-firefox` under temp Root

```
# Real-source path (temp ShellRoot)
stageFirefoxPublic(ShellRoot)
  -> BuildFirefoxExtensionShell(ShellRoot)
  -> StageFirefoxExtensionEmbed / Bundle dual-stage
  -> {ShellRoot|BundleRoot}/browseragent/embedded/extension-firefox/manifest.json

# Fixture path (Bundle UseFixture)
Bundle(UseFixture) with session fixture under BundleRoot
  -> may stage fixtures/extension-firefox into embed rel
```

## Preconditions

- Mode = `bundle-stages-firefox-embed`.
- Isolated `BundleRoot` and `ShellRoot` under `t.TempDir()` (no ModuleRoot embed write).
- Firefox public staged under ShellRoot.
- Session-page fixture path from ModuleRoot for Bundle UseFixture.
- Firefox fixture dir recorded for implementer discovery (`fixtures/extension-firefox`).

## Steps

1. Set `Mode = ModeBundleFirefox`.
2. Allocate BundleRoot + ShellRoot.
3. Stage minimal Firefox-Ext public under ShellRoot.
4. Point FixtureSessionPageDir / FixtureFirefoxExtensionDir at ModuleRoot fixtures.
5. Set UseFixture = true so Run also invokes Bundle.

## Context

- RED until Bundle dual-stages firefox embed and/or StageFirefoxExtensionEmbed is wired
  so manifest.json appears under `browseragent/embedded/extension-firefox`.
- BuildFirefoxExtensionShell alone (public→build) is **not** sufficient — embed path required.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeBundleFirefox
	req.UseFixture = true

	base := t.TempDir()
	req.BundleRoot = filepath.Join(base, "bundle-root")
	req.ShellRoot = filepath.Join(base, "shell-root")
	if err := os.MkdirAll(req.BundleRoot, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.ShellRoot, 0o755); err != nil {
		return err
	}

	// Real-source package for BuildFirefoxExtensionShell.
	stageFirefoxPublic(t, req.ShellRoot)

	// Fixtures for Bundle UseFixture + implementer firefox fixture discovery.
	mod := req.ModuleRoot
	req.FixtureSessionPageDir = filepath.Join(mod, "browseragent", "fixtures", "session-page")
	req.FixtureFirefoxExtensionDir = filepath.Join(mod, "browseragent", "fixtures", "extension-firefox")

	// Make firefox fixture visible under BundleRoot for implementers that resolve
	// fixtures relative to Root.
	dstFF := filepath.Join(req.BundleRoot, "browseragent", "fixtures", "extension-firefox")
	if err := copyTreeIfExists(t, req.FixtureFirefoxExtensionDir, dstFF); err != nil {
		// Soft: fixture may be absent in sparse checkouts; Stage still required via ShellRoot.
		t.Logf("copy firefox fixture: %v", err)
	}
	dstSess := filepath.Join(req.BundleRoot, "browseragent", "fixtures", "session-page")
	if err := copyTreeIfExists(t, req.FixtureSessionPageDir, dstSess); err != nil {
		t.Logf("copy session-page fixture: %v", err)
	}
	// Prefer absolute fixture paths (Bundle resolves overrides when absolute).
	if fileExists(filepath.Join(req.FixtureSessionPageDir, "index.html")) ||
		fileExists(filepath.Join(req.FixtureSessionPageDir, "session-page.html")) {
		// keep ModuleRoot absolute paths
	} else if fileExists(filepath.Join(dstSess, "index.html")) {
		req.FixtureSessionPageDir = dstSess
	}

	return nil
}
```
