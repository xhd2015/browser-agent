# Scenario

**Feature**: product shell public manifest is a production source under contract

```
ModuleRoot/Chrome-Ext-Browser-Agent/public/manifest.json
  -> ValidateExtensionManifestJSON
```

## Preconditions

- ManifestSource = shell.
- Prefer `public/manifest.json` (build source).

## Steps

1. Set ManifestSource to shell.
2. Leaf asserts Validate OK.

## Context

- Requirement A5. Shell must not drift from embedded permissions.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ManifestSource = ManifestSourceShell
	return nil
}
```
