# Scenario

**Feature**: embedded extension manifest satisfies required perms + hosts (A4)

```
browseragent/embedded/extension/manifest.json
  -> ValidateExtensionManifestJSON -> nil
  (debugger, tabs, alarms, storage, 43761, broad host)
```

## Preconditions

- Parent sets ManifestSource=embedded.
- ModuleRoot resolved at root Setup.

## Steps

1. No extra fields; Run resolves path from ModuleRoot.

## Context

- Requirement A4.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ManifestSource = ManifestSourceEmbedded
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	// Document expected path (Run also resolves it).
	req.ManifestPath = filepath.Join(req.ModuleRoot, "browseragent", "embedded", "extension", "manifest.json")
	return nil
}
```
