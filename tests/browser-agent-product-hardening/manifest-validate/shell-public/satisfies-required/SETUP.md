# Scenario

**Feature**: Chrome-Ext-Browser-Agent/public/manifest.json satisfies required perms + hosts (A5)

```
Chrome-Ext-Browser-Agent/public/manifest.json
  -> ValidateExtensionManifestJSON -> nil
  (debugger, tabs, alarms, storage, 43761, broad host)
```

## Preconditions

- Parent sets ManifestSource=shell.
- ModuleRoot resolved at root Setup.

## Steps

1. Prefer explicit public path; Run falls back to other shell locations if needed.

## Context

- Requirement A5.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ManifestSource = ManifestSourceShell
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	req.ManifestPath = filepath.Join(req.ModuleRoot, "Chrome-Ext-Browser-Agent", "public", "manifest.json")
	return nil
}
```
